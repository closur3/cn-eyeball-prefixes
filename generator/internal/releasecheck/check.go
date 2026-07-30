package releasecheck

import (
	"bufio"
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

type Options struct {
	CurrentRoot               string
	CandidateRoot             string
	MaxRemovedAddressRatio    float64
	MaxAddedAddressRatio      float64
	MaxPrefixCountGrowthRatio float64
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
	maxRemoved := flags.Float64("max-removed-address-ratio", 0.10, "removed-address warning threshold per public list")
	maxAdded := flags.Float64("max-added-address-ratio", 0.10, "added-address warning threshold per public list")
	maxPrefixGrowth := flags.Float64("max-prefix-count-growth-ratio", 3.0, "prefix-count growth warning threshold per public list")
	if err := flags.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	warnings, err := Check(Options{
		CurrentRoot:               *currentRoot,
		CandidateRoot:             *candidateRoot,
		MaxRemovedAddressRatio:    *maxRemoved,
		MaxAddedAddressRatio:      *maxAdded,
		MaxPrefixCountGrowthRatio: *maxPrefixGrowth,
	})
	if err != nil {
		panic(err)
	}
	if len(warnings) == 0 {
		fmt.Println("OK: candidate list changes are within warning thresholds")
		return
	}
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, "WARNING:", warning)
	}
}

func Check(options Options) ([]string, error) {
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

	var warnings []string
	for _, relative := range listmanifest.ExpectedPaths() {
		familyBits := 128
		if strings.HasPrefix(relative, "ipv4/") {
			familyBits = 32
		}
		current, err := readPrefixSet(
			filepath.Join(options.CurrentRoot, filepath.FromSlash(relative)),
			familyBits,
		)
		if err != nil {
			return nil, fmt.Errorf("current %s: %w", relative, err)
		}
		candidate, err := readPrefixSet(
			filepath.Join(options.CandidateRoot, filepath.FromSlash(relative)),
			familyBits,
		)
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

		if currentCount.Sign() == 0 {
			if candidateCount.Sign() != 0 {
				warnings = append(warnings, fmt.Sprintf(
					"%s has no historical baseline and is now non-empty",
					relative,
				))
			}
		} else {
			removedRatio := fraction(removed, currentCount)
			addedRatio := fraction(added, currentCount)
			if removedRatio > options.MaxRemovedAddressRatio {
				warnings = append(warnings, fmt.Sprintf(
					"%s removes %.2f%% of addresses",
					relative,
					removedRatio*100,
				))
			}
			if addedRatio > options.MaxAddedAddressRatio {
				warnings = append(warnings, fmt.Sprintf(
					"%s adds %.2f%% of addresses",
					relative,
					addedRatio*100,
				))
			}
		}
		if current.prefixCount != 0 {
			growth := float64(candidate.prefixCount) / float64(current.prefixCount)
			if growth > options.MaxPrefixCountGrowthRatio {
				warnings = append(warnings, fmt.Sprintf(
					"%s prefix count grows %.2fx",
					relative,
					growth,
				))
			}
		}
	}
	return warnings, nil
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
		line := scanner.Text()
		if strings.TrimSpace(line) != line || line == "" {
			return prefixSet{}, fmt.Errorf("non-canonical line %d", result.prefixCount+1)
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
			return prefixSet{}, fmt.Errorf("CIDRs are not sorted and disjoint at line %d", result.prefixCount+1)
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
	for _, item := range intervals {
		size := new(big.Int).Sub(item.hi, item.lo)
		total.Add(total, size.Add(size, big.NewInt(1)))
	}
	return total
}

func intersectionCount(left, right []interval) *big.Int {
	total := new(big.Int)
	for leftIndex, rightIndex := 0, 0; leftIndex < len(left) && rightIndex < len(right); {
		lo := left[leftIndex].lo
		if right[rightIndex].lo.Cmp(lo) > 0 {
			lo = right[rightIndex].lo
		}
		hi := left[leftIndex].hi
		if right[rightIndex].hi.Cmp(hi) < 0 {
			hi = right[rightIndex].hi
		}
		if lo.Cmp(hi) <= 0 {
			size := new(big.Int).Sub(hi, lo)
			total.Add(total, size.Add(size, big.NewInt(1)))
		}
		if left[leftIndex].hi.Cmp(right[rightIndex].hi) < 0 {
			leftIndex++
		} else {
			rightIndex++
		}
	}
	return total
}

func fraction(numerator, denominator *big.Int) float64 {
	value, _ := new(big.Rat).SetFrac(numerator, denominator).Float64()
	return value
}

func isRequiredList(relative string) bool {
	return strings.Count(filepath.ToSlash(relative), "/") == 1
}
