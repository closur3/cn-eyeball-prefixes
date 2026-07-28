package releasecheck

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/big"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/listmanifest"
)

const ReportSchemaVersion = 3

type Options struct {
	CurrentRoot               string
	CandidateRoot             string
	ReportPath                string
	MaxRemovedAddressRatio    float64
	MaxAddedAddressRatio      float64
	MaxPrefixCountGrowthRatio float64
}

type Thresholds struct {
	MaxRemovedAddressRatio    float64 `json:"max_removed_address_ratio"`
	MaxAddedAddressRatio      float64 `json:"max_added_address_ratio"`
	MaxPrefixCountGrowthRatio float64 `json:"max_prefix_count_growth_ratio"`
}

type FileReport struct {
	Path                   string   `json:"path"`
	CurrentPrefixCount     int      `json:"current_prefix_count"`
	CandidatePrefixCount   int      `json:"candidate_prefix_count"`
	CurrentAddressCount    string   `json:"current_address_count"`
	CandidateAddressCount  string   `json:"candidate_address_count"`
	RemovedAddressCount    string   `json:"removed_address_count"`
	AddedAddressCount      string   `json:"added_address_count"`
	BaselineEmpty          bool     `json:"baseline_empty"`
	RemovedAddressRatio    *float64 `json:"removed_address_ratio"`
	AddedAddressRatio      *float64 `json:"added_address_ratio"`
	PrefixCountGrowthRatio *float64 `json:"prefix_count_growth_ratio"`
	JaccardSimilarity      float64  `json:"jaccard_similarity"`
}

type Report struct {
	SchemaVersion      int          `json:"schema_version"`
	CandidateContentID string       `json:"candidate_content_id"`
	Thresholds         Thresholds   `json:"thresholds"`
	Files              []FileReport `json:"files"`
	Warnings           []string     `json:"warnings"`
}

type interval struct {
	lo *big.Int
	hi *big.Int
}

type prefixSet struct {
	prefixCount int
	intervals   []interval
}

func Main() {
	flags := flag.NewFlagSet("verify release", flag.ExitOnError)
	currentRoot := flags.String("current", "", "current public lists root")
	candidateRoot := flags.String("candidate", "", "candidate public lists root")
	reportPath := flags.String("report", "", "optional JSON report path")
	maxRemoved := flags.Float64("max-removed-address-ratio", 0.10, "removed-address warning threshold per public list")
	maxAdded := flags.Float64("max-added-address-ratio", 0.10, "added-address warning threshold per public list")
	maxPrefixGrowth := flags.Float64("max-prefix-count-growth-ratio", 3.0, "prefix-count growth warning threshold per public list")
	if err := flags.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	report, err := Check(Options{
		CurrentRoot:               *currentRoot,
		CandidateRoot:             *candidateRoot,
		ReportPath:                *reportPath,
		MaxRemovedAddressRatio:    *maxRemoved,
		MaxAddedAddressRatio:      *maxAdded,
		MaxPrefixCountGrowthRatio: *maxPrefixGrowth,
	})
	if err != nil {
		panic(err)
	}

	for _, entry := range report.Files {
		if entry.CurrentPrefixCount == entry.CandidatePrefixCount &&
			entry.RemovedAddressCount == "0" &&
			entry.AddedAddressCount == "0" {
			continue
		}
		fmt.Printf(
			"%s: prefixes %d -> %d; removed %s; added %s; jaccard %.6f\n",
			entry.Path,
			entry.CurrentPrefixCount,
			entry.CandidatePrefixCount,
			formatRatio(entry.RemovedAddressRatio),
			formatRatio(entry.AddedAddressRatio),
			entry.JaccardSimilarity,
		)
	}

	if len(report.Warnings) != 0 {
		for _, warning := range report.Warnings {
			fmt.Fprintln(os.Stderr, "release warning:", warning)
		}
		fmt.Println("WARNING: candidate list changes exceed release review thresholds")
		return
	}
	fmt.Println("OK: candidate list changes are within release review thresholds")
}

func Check(options Options) (*Report, error) {
	if options.CurrentRoot == "" {
		return nil, fmt.Errorf("--current is required")
	}
	if options.CandidateRoot == "" {
		return nil, fmt.Errorf("--candidate is required")
	}
	if err := validateRatio("max removed address ratio", options.MaxRemovedAddressRatio, 0, 1); err != nil {
		return nil, err
	}
	if err := validateRatio("max added address ratio", options.MaxAddedAddressRatio, 0, 0); err != nil {
		return nil, err
	}
	if err := validateRatio("max prefix-count growth ratio", options.MaxPrefixCountGrowthRatio, 1, 0); err != nil {
		return nil, err
	}
	candidateContentID, err := listmanifest.ComputeContentID(options.CandidateRoot)
	if err != nil {
		return nil, fmt.Errorf("candidate content_id: %w", err)
	}

	report := &Report{
		SchemaVersion:      ReportSchemaVersion,
		CandidateContentID: candidateContentID,
		Warnings:           []string{},
		Thresholds: Thresholds{
			MaxRemovedAddressRatio:    options.MaxRemovedAddressRatio,
			MaxAddedAddressRatio:      options.MaxAddedAddressRatio,
			MaxPrefixCountGrowthRatio: options.MaxPrefixCountGrowthRatio,
		},
	}
	for _, relative := range listmanifest.ExpectedPaths() {
		familyBits := 128
		if strings.HasPrefix(relative, "ipv4/") {
			familyBits = 32
		}
		current, err := readPrefixSet(filepath.Join(options.CurrentRoot, filepath.FromSlash(relative)), familyBits)
		if err != nil {
			return nil, fmt.Errorf("current %s: %w", relative, err)
		}
		candidate, err := readPrefixSet(filepath.Join(options.CandidateRoot, filepath.FromSlash(relative)), familyBits)
		if err != nil {
			return nil, fmt.Errorf("candidate %s: %w", relative, err)
		}

		if isRequiredList(relative) && candidate.prefixCount == 0 {
			return nil, fmt.Errorf("candidate %s is empty", relative)
		}

		currentCount := addressCount(current.intervals)
		candidateCount := addressCount(candidate.intervals)
		intersection := intersectionCount(current.intervals, candidate.intervals)
		removed := new(big.Int).Sub(new(big.Int).Set(currentCount), intersection)
		added := new(big.Int).Sub(new(big.Int).Set(candidateCount), intersection)
		union := new(big.Int).Sub(new(big.Int).Add(new(big.Int).Set(currentCount), candidateCount), intersection)

		baselineEmpty := currentCount.Sign() == 0
		removedRatio := optionalFraction(removed, currentCount)
		addedRatio := optionalFraction(added, currentCount)
		prefixGrowth := optionalCountRatio(candidate.prefixCount, current.prefixCount)
		jaccard := 1.0
		if union.Sign() != 0 {
			jaccard = fraction(intersection, union)
		}
		entry := FileReport{
			Path:                   relative,
			CurrentPrefixCount:     current.prefixCount,
			CandidatePrefixCount:   candidate.prefixCount,
			CurrentAddressCount:    currentCount.String(),
			CandidateAddressCount:  candidateCount.String(),
			RemovedAddressCount:    removed.String(),
			AddedAddressCount:      added.String(),
			BaselineEmpty:          baselineEmpty,
			RemovedAddressRatio:    removedRatio,
			AddedAddressRatio:      addedRatio,
			PrefixCountGrowthRatio: prefixGrowth,
			JaccardSimilarity:      jaccard,
		}
		report.Files = append(report.Files, entry)

		if baselineEmpty && candidateCount.Sign() != 0 {
			report.Warnings = append(report.Warnings, fmt.Sprintf(
				"%s has no historical address baseline; candidate contains %s addresses",
				relative,
				candidateCount,
			))
		}
		if removedRatio != nil && *removedRatio > options.MaxRemovedAddressRatio {
			report.Warnings = append(report.Warnings, fmt.Sprintf(
				"%s removes %.6f of the current address set (limit %.6f)",
				relative,
				*removedRatio,
				options.MaxRemovedAddressRatio,
			))
		}
		if addedRatio != nil && *addedRatio > options.MaxAddedAddressRatio {
			report.Warnings = append(report.Warnings, fmt.Sprintf(
				"%s adds %.6f of the current address set (limit %.6f)",
				relative,
				*addedRatio,
				options.MaxAddedAddressRatio,
			))
		}
		if prefixGrowth != nil && *prefixGrowth > options.MaxPrefixCountGrowthRatio {
			report.Warnings = append(report.Warnings, fmt.Sprintf(
				"%s grows from %d to %d prefixes (%.3fx; limit %.3fx)",
				relative,
				current.prefixCount,
				candidate.prefixCount,
				*prefixGrowth,
				options.MaxPrefixCountGrowthRatio,
			))
		}
	}
	if options.ReportPath != "" {
		if err := writeReport(options.ReportPath, report); err != nil {
			return nil, err
		}
	}
	return report, nil
}

func validateRatio(name string, value, minimum, maximum float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("%s must be finite", name)
	}
	if value < minimum || (maximum != 0 && value > maximum) {
		if maximum == 0 {
			return fmt.Errorf("%s must be at least %.1f", name, minimum)
		}
		return fmt.Errorf("%s must be between %.1f and %.1f", name, minimum, maximum)
	}
	return nil
}

func readPrefixSet(path string, familyBits int) (prefixSet, error) {
	file, err := os.Open(path)
	if err != nil {
		return prefixSet{}, err
	}
	defer file.Close()

	var result prefixSet
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return prefixSet{}, fmt.Errorf("blank line at line %d", result.prefixCount+1)
		}
		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return prefixSet{}, fmt.Errorf("invalid CIDR at line %d: %w", result.prefixCount+1, err)
		}
		if prefix != prefix.Masked() || prefix.String() != line {
			return prefixSet{}, fmt.Errorf("non-canonical CIDR at line %d: %s", result.prefixCount+1, line)
		}
		ipv4 := familyBits == 32
		if ipv4 != prefix.Addr().Is4() ||
			(!ipv4 && (!prefix.Addr().Is6() || prefix.Addr().Is4In6())) {
			return prefixSet{}, fmt.Errorf("wrong address family at line %d: %s", result.prefixCount+1, line)
		}

		lo := addressInteger(prefix.Addr())
		size := new(big.Int).Lsh(big.NewInt(1), uint(familyBits-prefix.Bits()))
		hi := new(big.Int).Sub(new(big.Int).Add(new(big.Int).Set(lo), size), big.NewInt(1))
		if len(result.intervals) != 0 && lo.Cmp(result.intervals[len(result.intervals)-1].hi) <= 0 {
			return prefixSet{}, fmt.Errorf("CIDRs are not strictly sorted and disjoint at line %d", result.prefixCount+1)
		}
		result.intervals = append(result.intervals, interval{lo: lo, hi: hi})
		result.prefixCount++
	}
	if err := scanner.Err(); err != nil {
		return prefixSet{}, err
	}
	return result, nil
}

func addressInteger(address netip.Addr) *big.Int {
	if address.Is4() {
		value := address.As4()
		return new(big.Int).SetBytes(value[:])
	}
	value := address.As16()
	return new(big.Int).SetBytes(value[:])
}

func addressCount(intervals []interval) *big.Int {
	total := new(big.Int)
	one := big.NewInt(1)
	for _, item := range intervals {
		size := new(big.Int).Add(new(big.Int).Sub(item.hi, item.lo), one)
		total.Add(total, size)
	}
	return total
}

func intersectionCount(left, right []interval) *big.Int {
	total := new(big.Int)
	one := big.NewInt(1)
	for i, j := 0, 0; i < len(left) && j < len(right); {
		lo := left[i].lo
		if right[j].lo.Cmp(lo) > 0 {
			lo = right[j].lo
		}
		hi := left[i].hi
		if right[j].hi.Cmp(hi) < 0 {
			hi = right[j].hi
		}
		if lo.Cmp(hi) <= 0 {
			size := new(big.Int).Add(new(big.Int).Sub(hi, lo), one)
			total.Add(total, size)
		}
		if left[i].hi.Cmp(right[j].hi) < 0 {
			i++
		} else {
			j++
		}
	}
	return total
}

func fraction(numerator, denominator *big.Int) float64 {
	value, _ := new(big.Rat).SetFrac(numerator, denominator).Float64()
	return value
}

func optionalFraction(numerator, denominator *big.Int) *float64 {
	if denominator.Sign() == 0 {
		return nil
	}
	value := fraction(numerator, denominator)
	return &value
}

func optionalCountRatio(numerator, denominator int) *float64 {
	if denominator == 0 {
		return nil
	}
	value := float64(numerator) / float64(denominator)
	return &value
}

func formatRatio(value *float64) string {
	if value == nil {
		return "n/a"
	}
	return fmt.Sprintf("%.6f", *value)
}

func isRequiredList(relative string) bool {
	return strings.Count(filepath.ToSlash(relative), "/") == 1
}

func writeReport(path string, report *Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	return nil
}
