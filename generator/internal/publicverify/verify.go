// Package publicverify checks relationships that can be proven from the
// published list files alone, without re-downloading any upstream data.
package publicverify

import (
	"bufio"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/iputil"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/listmanifest"
)

const expectedProvinceCount = 31

type addressRange struct {
	first netip.Addr
	last  netip.Addr
}

// Verify checks the operator and province relationships for both address
// families under root. It reads only the public TXT files.
func Verify(root string) error {
	if strings.TrimSpace(root) == "" {
		return fmt.Errorf("public list root is required")
	}
	for _, family := range []string{"ipv4", "ipv6"} {
		if err := verifyFamily(root, family); err != nil {
			return err
		}
	}
	return nil
}

func verifyFamily(root, family string) error {
	ipv4 := family == "ipv4"
	cn, err := readRelative(root, family+"/cn.txt", ipv4)
	if err != nil {
		return err
	}

	operators := append([]string(nil), iputil.Operators...)
	byOperator := make(map[string][]addressRange, len(operators))
	var operatorUnion []addressRange
	for _, operator := range operators {
		relative := family + "/" + operator + ".txt"
		ranges, err := readRelative(root, relative, ipv4)
		if err != nil {
			return err
		}
		if len(ranges) == 0 {
			return fmt.Errorf("%s contains no prefixes", relative)
		}
		byOperator[operator] = ranges
		operatorUnion = append(operatorUnion, ranges...)
	}
	if err := rejectPairwiseOverlaps(family+" operator", operators, byOperator); err != nil {
		return err
	}
	if !equalSets(operatorUnion, cn) {
		return fmt.Errorf("%s operator union does not equal %s/cn.txt", family, family)
	}

	provinces, err := provinceNames(family)
	if err != nil {
		return err
	}
	byProvince := make(map[string][]addressRange, len(provinces))
	var provinceUnion []addressRange
	for _, province := range provinces {
		relative := family + "/provinces/" + province + ".txt"
		ranges, err := readRelative(root, relative, ipv4)
		if err != nil {
			return err
		}
		if ok, outside := containedBy(cn, ranges); !ok {
			return fmt.Errorf(
				"%s contains range %s-%s outside %s/cn.txt",
				relative,
				outside.first,
				outside.last,
				family,
			)
		}
		byProvince[province] = ranges
		provinceUnion = append(provinceUnion, ranges...)
	}
	if err := rejectPairwiseOverlaps(family+" province", provinces, byProvince); err != nil {
		return err
	}
	if ipv4 {
		provinceAddressCount := ipv4AddressCount(normalize(provinceUnion))
		cnAddressCount := ipv4AddressCount(normalize(cn))
		if provinceAddressCount*100 < cnAddressCount*90 {
			return fmt.Errorf(
				"ipv4 province lists cover fewer than 90%% of ipv4/cn.txt addresses",
			)
		}
	} else if !equalSets(provinceUnion, cn) {
		return fmt.Errorf("ipv6 province union does not equal ipv6/cn.txt")
	}
	return nil
}

func ipv4AddressCount(ranges []addressRange) uint64 {
	var total uint64
	for _, current := range ranges {
		first := current.first.As4()
		last := current.last.As4()
		firstInteger := uint64(first[0])<<24 |
			uint64(first[1])<<16 |
			uint64(first[2])<<8 |
			uint64(first[3])
		lastInteger := uint64(last[0])<<24 |
			uint64(last[1])<<16 |
			uint64(last[2])<<8 |
			uint64(last[3])
		total += lastInteger - firstInteger + 1
	}
	return total
}

func provinceNames(family string) ([]string, error) {
	prefix := family + "/provinces/"
	var provinces []string
	for _, relative := range listmanifest.ExpectedPaths() {
		if !strings.HasPrefix(relative, prefix) || !strings.HasSuffix(relative, ".txt") {
			continue
		}
		province := strings.TrimSuffix(strings.TrimPrefix(relative, prefix), ".txt")
		if province == "" || strings.Contains(province, "/") {
			return nil, fmt.Errorf("invalid public province path %q", relative)
		}
		provinces = append(provinces, province)
	}
	if len(provinces) != expectedProvinceCount {
		return nil, fmt.Errorf(
			"public path contract has %d %s provinces, want %d",
			len(provinces),
			family,
			expectedProvinceCount,
		)
	}
	return provinces, nil
}

func readRelative(root, relative string, ipv4 bool) ([]addressRange, error) {
	path := filepath.Join(root, filepath.FromSlash(relative))
	ranges, err := readPrefixFile(path, ipv4)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", relative, err)
	}
	return ranges, nil
}

func readPrefixFile(path string, ipv4 bool) ([]addressRange, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ranges []addressRange
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			return nil, fmt.Errorf("blank line at line %d", lineNumber)
		}
		if strings.TrimSpace(line) != line {
			return nil, fmt.Errorf("non-canonical CIDR text at line %d", lineNumber)
		}
		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR at line %d: %w", lineNumber, err)
		}
		if prefix != prefix.Masked() || prefix.String() != line {
			return nil, fmt.Errorf("non-canonical CIDR at line %d: %s", lineNumber, line)
		}
		if ipv4 != prefix.Addr().Is4() || (!ipv4 && (!prefix.Addr().Is6() || prefix.Addr().Is4In6())) {
			return nil, fmt.Errorf("wrong address family at line %d: %s", lineNumber, line)
		}

		current := addressRange{first: prefix.Addr(), last: lastAddress(prefix)}
		if len(ranges) != 0 {
			previous := ranges[len(ranges)-1]
			if previous.first.Compare(current.first) >= 0 {
				return nil, fmt.Errorf("CIDRs are not strictly sorted at line %d", lineNumber)
			}
			if current.first.Compare(previous.last) <= 0 {
				return nil, fmt.Errorf("CIDRs overlap at line %d", lineNumber)
			}
		}
		ranges = append(ranges, current)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ranges, nil
}

func lastAddress(prefix netip.Prefix) netip.Addr {
	if prefix.Addr().Is4() {
		value := prefix.Addr().As4()
		setHostBits(value[:], prefix.Bits())
		return netip.AddrFrom4(value)
	}
	value := prefix.Addr().As16()
	setHostBits(value[:], prefix.Bits())
	return netip.AddrFrom16(value)
}

func setHostBits(value []byte, prefixBits int) {
	for bit := prefixBits; bit < len(value)*8; bit++ {
		value[bit/8] |= 1 << uint(7-bit%8)
	}
}

func rejectPairwiseOverlaps(kind string, names []string, sets map[string][]addressRange) error {
	for leftIndex, leftName := range names {
		for _, rightName := range names[leftIndex+1:] {
			left, right, ok := firstOverlap(sets[leftName], sets[rightName])
			if ok {
				return fmt.Errorf(
					"%s lists overlap: %s %s-%s and %s %s-%s",
					kind,
					leftName,
					left.first,
					left.last,
					rightName,
					right.first,
					right.last,
				)
			}
		}
	}
	return nil
}

func firstOverlap(left, right []addressRange) (addressRange, addressRange, bool) {
	for leftIndex, rightIndex := 0, 0; leftIndex < len(left) && rightIndex < len(right); {
		leftRange := left[leftIndex]
		rightRange := right[rightIndex]
		switch {
		case leftRange.last.Compare(rightRange.first) < 0:
			leftIndex++
		case rightRange.last.Compare(leftRange.first) < 0:
			rightIndex++
		default:
			return leftRange, rightRange, true
		}
	}
	return addressRange{}, addressRange{}, false
}

func containedBy(container, candidate []addressRange) (bool, addressRange) {
	container = normalize(container)
	candidate = normalize(candidate)
	containerIndex := 0
	for _, current := range candidate {
		for containerIndex < len(container) &&
			container[containerIndex].last.Compare(current.first) < 0 {
			containerIndex++
		}
		if containerIndex == len(container) ||
			container[containerIndex].first.Compare(current.first) > 0 ||
			container[containerIndex].last.Compare(current.last) < 0 {
			return false, current
		}
	}
	return true, addressRange{}
}

func equalSets(left, right []addressRange) bool {
	left = normalize(left)
	right = normalize(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func normalize(values []addressRange) []addressRange {
	ordered := append([]addressRange(nil), values...)
	sort.Slice(ordered, func(left, right int) bool {
		if comparison := ordered[left].first.Compare(ordered[right].first); comparison != 0 {
			return comparison < 0
		}
		return ordered[left].last.Compare(ordered[right].last) < 0
	})

	merged := make([]addressRange, 0, len(ordered))
	for _, current := range ordered {
		if len(merged) == 0 {
			merged = append(merged, current)
			continue
		}
		previous := &merged[len(merged)-1]
		overlaps := current.first.Compare(previous.last) <= 0
		adjacent := false
		if !overlaps {
			next := previous.last.Next()
			adjacent = next.IsValid() && current.first == next
		}
		if !overlaps && !adjacent {
			merged = append(merged, current)
			continue
		}
		if current.last.Compare(previous.last) > 0 {
			previous.last = current.last
		}
	}
	return merged
}
