package operatorconfig

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type operator struct {
	DescriptionPatterns []string          `json:"description_patterns"`
	IncludeASNs         map[string]string `json:"include_asns"`
}

type descriptionRule struct {
	Pattern string `json:"pattern"`
	Reason  string `json:"reason"`
	Inherit bool   `json:"inherit,omitempty"`
	Hard    bool   `json:"hard,omitempty"`
}

type configFile struct {
	Operators                      map[string]operator `json:"operators"`
	ExcludeDescriptionRules        []descriptionRule   `json:"exclude_description_rules"`
	ExcludeAPNICInetnumRules       []descriptionRule   `json:"exclude_apnic_inetnum_rules"`
	AccessAPNICInetnumRules        []descriptionRule   `json:"access_apnic_inetnum_rules"`
	ExcludeBackboneIPv4Prefixes    []BackboneIPv4Prefix `json:"exclude_backbone_ipv4_prefixes"`
	ExcludeASNs                    map[string]string   `json:"exclude_asns"`
	IndependentLegalEntityPatterns []string            `json:"independent_legal_entity_patterns"`
}

type rule struct {
	name     string
	patterns []matchPattern
}

type matchPattern struct {
	pattern *regexp.Regexp
	source  string
}

type exclusionRule struct {
	pattern *regexp.Regexp
	source  string
	reason  string
	inherit bool
	hard    bool
}

type Result struct {
	Operator        string
	Excluded        bool
	Reason          string
	MatchedBy       string
	ExclusionSource string
}

type PrefixResult struct {
	Excluded  bool
	Reason    string
	MatchedBy string
	Inherit   bool
	Hard      bool
}

type BackboneIPv4Prefix struct {
	CIDR         string   `json:"cidr"`
	Operator     string   `json:"operator"`
	Reason       string   `json:"reason"`
	EvidenceURLs []string `json:"evidence_urls"`
}

type inclusion struct {
	operator string
	reason   string
}

type Classifier struct {
	rules                        []rule
	included                     map[string]inclusion
	excluded                     map[string]string
	exclusionPatterns            []exclusionRule
	apnicPatterns                []exclusionRule
	apnicAccessPatterns          []exclusionRule
	reviewedBackboneIPv4Prefixes []BackboneIPv4Prefix
	legalEntityPatterns          []*regexp.Regexp
}

func Load(path string, order []string) (*Classifier, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(b, order)
}

func Parse(b []byte, order []string) (*Classifier, error) {
	var cfg configFile
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse operator config: %w", err)
	}
	if len(cfg.Operators) != len(order) {
		return nil, fmt.Errorf("operator config has %d operators, want %d", len(cfg.Operators), len(order))
	}
	c := &Classifier{included: map[string]inclusion{}, excluded: map[string]string{}}
	for asn, reason := range cfg.ExcludeASNs {
		if err := validASN(asn); err != nil {
			return nil, fmt.Errorf("excluded ASN %q: %w", asn, err)
		}
		if reason == "" {
			return nil, fmt.Errorf("excluded ASN %q has no reason", asn)
		}
		c.excluded[asn] = reason
	}
	for _, rule := range cfg.ExcludeDescriptionRules {
		if rule.Pattern == "" || rule.Reason == "" {
			return nil, fmt.Errorf("exclude description rules require both pattern and reason")
		}
		re, err := regexp.Compile("(?i)(?:" + rule.Pattern + ")")
		if err != nil {
			return nil, fmt.Errorf("exclude description pattern %q: %w", rule.Pattern, err)
		}
		c.exclusionPatterns = append(c.exclusionPatterns, exclusionRule{pattern: re, source: rule.Pattern, reason: rule.Reason})
	}
	if len(cfg.ExcludeAPNICInetnumRules) == 0 {
		return nil, fmt.Errorf("operator config has no APNIC inetnum exclusion rules")
	}
	for _, rule := range cfg.ExcludeAPNICInetnumRules {
		if rule.Pattern == "" || rule.Reason == "" {
			return nil, fmt.Errorf("APNIC inetnum exclusion rules require both pattern and reason")
		}
		if rule.Hard && !rule.Inherit {
			return nil, fmt.Errorf("hard APNIC inetnum exclusion rule %q must also inherit", rule.Pattern)
		}
		re, err := regexp.Compile("(?i)(?:" + rule.Pattern + ")")
		if err != nil {
			return nil, fmt.Errorf("APNIC inetnum exclusion pattern %q: %w", rule.Pattern, err)
		}
		c.apnicPatterns = append(c.apnicPatterns, exclusionRule{
			pattern: re,
			source:  rule.Pattern,
			reason:  rule.Reason,
			inherit: rule.Inherit,
			hard:    rule.Hard,
		})
	}
	if len(cfg.AccessAPNICInetnumRules) == 0 {
		return nil, fmt.Errorf("operator config has no APNIC inetnum access rules")
	}
	for _, rule := range cfg.AccessAPNICInetnumRules {
		if rule.Pattern == "" || rule.Reason == "" {
			return nil, fmt.Errorf("APNIC inetnum access rules require both pattern and reason")
		}
		if rule.Inherit || rule.Hard {
			return nil, fmt.Errorf("APNIC inetnum access rule %q cannot set exclusion inheritance", rule.Pattern)
		}
		re, err := regexp.Compile("(?i)(?:" + rule.Pattern + ")")
		if err != nil {
			return nil, fmt.Errorf("APNIC inetnum access pattern %q: %w", rule.Pattern, err)
		}
		c.apnicAccessPatterns = append(c.apnicAccessPatterns, exclusionRule{pattern: re, source: rule.Pattern, reason: rule.Reason})
	}
	if len(cfg.IndependentLegalEntityPatterns) == 0 {
		return nil, fmt.Errorf("operator config has no independent legal-entity patterns")
	}
	for _, pattern := range cfg.IndependentLegalEntityPatterns {
		re, err := regexp.Compile("(?i)(?:" + pattern + ")")
		if err != nil {
			return nil, fmt.Errorf("independent legal-entity pattern %q: %w", pattern, err)
		}
		c.legalEntityPatterns = append(c.legalEntityPatterns, re)
	}
	for _, name := range order {
		op, ok := cfg.Operators[name]
		if !ok {
			return nil, fmt.Errorf("operator config is missing %q", name)
		}
		if len(op.DescriptionPatterns) == 0 && len(op.IncludeASNs) == 0 {
			return nil, fmt.Errorf("operator %q has no matching rules", name)
		}
		r := rule{name: name}
		for _, pattern := range op.DescriptionPatterns {
			re, err := regexp.Compile("(?i)(?:" + pattern + ")")
			if err != nil {
				return nil, fmt.Errorf("operator %q pattern %q: %w", name, pattern, err)
			}
			r.patterns = append(r.patterns, matchPattern{pattern: re, source: pattern})
		}
		for asn, reason := range op.IncludeASNs {
			if err := validASN(asn); err != nil {
				return nil, fmt.Errorf("operator %q included ASN %q: %w", name, asn, err)
			}
			if c.excluded[asn] != "" {
				return nil, fmt.Errorf("ASN %s is both included and excluded", asn)
			}
			if reason == "" {
				return nil, fmt.Errorf("operator %q included ASN %q has no reason", name, asn)
			}
			if previous, exists := c.included[asn]; exists {
				return nil, fmt.Errorf("ASN %s is included by both %s and %s", asn, previous.operator, name)
			}
			c.included[asn] = inclusion{operator: name, reason: reason}
		}
		c.rules = append(c.rules, r)
	}
	backbonePrefixes, err := validateBackboneIPv4Prefixes(cfg.ExcludeBackboneIPv4Prefixes, cfg.Operators)
	if err != nil {
		return nil, err
	}
	c.reviewedBackboneIPv4Prefixes = backbonePrefixes
	return c, nil
}

func validateBackboneIPv4Prefixes(entries []BackboneIPv4Prefix, operators map[string]operator) ([]BackboneIPv4Prefix, error) {
	type parsedEntry struct {
		cidr   string
		prefix netip.Prefix
	}

	seen := make(map[string]struct{}, len(entries))
	parsed := make([]parsedEntry, 0, len(entries))
	result := make([]BackboneIPv4Prefix, 0, len(entries))
	for i, entry := range entries {
		if _, ok := operators[entry.Operator]; !ok {
			return nil, fmt.Errorf("backbone IPv4 prefix %q has unknown operator %q", entry.CIDR, entry.Operator)
		}
		if strings.TrimSpace(entry.Reason) == "" {
			return nil, fmt.Errorf("backbone IPv4 prefix %q has no reason", entry.CIDR)
		}
		if len(entry.EvidenceURLs) == 0 {
			return nil, fmt.Errorf("backbone IPv4 prefix %q has no evidence URLs", entry.CIDR)
		}
		for _, evidenceURL := range entry.EvidenceURLs {
			u, err := url.Parse(evidenceURL)
			if err != nil || u.Scheme != "https" || u.Host == "" {
				return nil, fmt.Errorf("backbone IPv4 prefix %q evidence URL %q must be an absolute HTTPS URL", entry.CIDR, evidenceURL)
			}
		}

		prefix, err := netip.ParsePrefix(entry.CIDR)
		if err != nil {
			return nil, fmt.Errorf("backbone IPv4 prefix entry %d has invalid CIDR %q: %w", i, entry.CIDR, err)
		}
		if !prefix.Addr().Is4() {
			return nil, fmt.Errorf("backbone IPv4 prefix %q is not IPv4", entry.CIDR)
		}
		if prefix != prefix.Masked() || entry.CIDR != prefix.String() {
			return nil, fmt.Errorf("backbone IPv4 prefix %q is not canonical", entry.CIDR)
		}
		if _, ok := seen[entry.CIDR]; ok {
			return nil, fmt.Errorf("duplicate backbone IPv4 prefix %q", entry.CIDR)
		}
		for _, previous := range parsed {
			if prefix.Contains(previous.prefix.Addr()) || previous.prefix.Contains(prefix.Addr()) {
				return nil, fmt.Errorf("backbone IPv4 prefixes %q and %q overlap", previous.cidr, entry.CIDR)
			}
		}
		seen[entry.CIDR] = struct{}{}
		parsed = append(parsed, parsedEntry{cidr: entry.CIDR, prefix: prefix})
		entry.EvidenceURLs = append([]string(nil), entry.EvidenceURLs...)
		result = append(result, entry)
	}
	return result, nil
}

func (c *Classifier) ReviewedBackboneIPv4Prefixes() []BackboneIPv4Prefix {
	result := make([]BackboneIPv4Prefix, len(c.reviewedBackboneIPv4Prefixes))
	for i, entry := range c.reviewedBackboneIPv4Prefixes {
		entry.EvidenceURLs = append([]string(nil), entry.EvidenceURLs...)
		result[i] = entry
	}
	return result
}

// ClassifyAPNICRegistrant positively attributes an APNIC inetnum registrant to
// an operator. ASN-only exceptions and ASN exclusion policy deliberately do
// not participate: this is evidence about the most-specific registration, not
// the BGP origin.
func (c *Classifier) ClassifyAPNICRegistrant(text string) Result {
	for _, r := range c.rules {
		for _, pattern := range r.patterns {
			if pattern.pattern.MatchString(text) {
				return Result{Operator: r.name, MatchedBy: "description_patterns: " + pattern.source}
			}
		}
	}
	return Result{}
}

func (c *Classifier) IsIndependentLegalEntity(text string) bool {
	for _, pattern := range c.legalEntityPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func validASN(asn string) error {
	n, err := strconv.ParseUint(asn, 10, 32)
	if err != nil || n == 0 {
		return fmt.Errorf("must be an unsigned 32-bit integer greater than zero")
	}
	return nil
}

func (c *Classifier) Match(asn, description string) string {
	result := c.Classify(asn, description)
	if result.Excluded {
		return ""
	}
	return result.Operator
}

func (c *Classifier) Classify(asn, description string) Result {
	operator := ""
	matchedBy := ""
	if entry, ok := c.included[asn]; ok {
		operator = entry.operator
		matchedBy = "include_asns: " + entry.reason
	} else {
		for _, r := range c.rules {
			for _, pattern := range r.patterns {
				if pattern.pattern.MatchString(description) {
					operator = r.name
					matchedBy = "description_patterns: " + pattern.source
					break
				}
			}
			if operator != "" {
				break
			}
		}
	}
	if operator == "" {
		return Result{}
	}
	if reason := c.excluded[asn]; reason != "" {
		return Result{Operator: operator, Excluded: true, Reason: reason, MatchedBy: matchedBy, ExclusionSource: "explicit_policy"}
	}
	for _, rule := range c.exclusionPatterns {
		if rule.pattern.MatchString(description) {
			return Result{Operator: operator, Excluded: true, Reason: rule.reason, MatchedBy: matchedBy, ExclusionSource: "description_rule"}
		}
	}
	return Result{Operator: operator, MatchedBy: matchedBy}
}

func (c *Classifier) ClassifyAPNICInetnum(text string) PrefixResult {
	// APNIC RPSL descriptions contain inconsistent runs of spaces and tabs.
	// Normalize whitespace before matching so strong-purpose phrases such as
	// "Data  Center" cannot bypass otherwise exact exclusion rules.
	return c.classifyAPNICInetnumNormalized(normalizeAPNICText(text))
}

func (c *Classifier) classifyAPNICInetnumNormalized(text string) PrefixResult {
	for _, rule := range c.apnicPatterns {
		if rule.pattern.MatchString(text) {
			return PrefixResult{
				Excluded:  true,
				Reason:    rule.reason,
				MatchedBy: "exclude_apnic_inetnum_rules: " + rule.source,
				Inherit:   rule.inherit,
				Hard:      rule.hard,
			}
		}
	}
	return PrefixResult{}
}

func (c *Classifier) ClassifyAPNICAccess(text string) PrefixResult {
	return c.classifyAPNICAccessNormalized(normalizeAPNICText(text))
}

func (c *Classifier) classifyAPNICAccessNormalized(text string) PrefixResult {
	for _, rule := range c.apnicAccessPatterns {
		if rule.pattern.MatchString(text) {
			return PrefixResult{Reason: rule.reason, MatchedBy: "access_apnic_inetnum_rules: " + rule.source}
		}
	}
	return PrefixResult{}
}

// ClassifyAPNICPolicy computes exclusion and terminal-access evidence after a
// single whitespace-normalization pass.
func (c *Classifier) ClassifyAPNICPolicy(text string) (PrefixResult, PrefixResult) {
	text = normalizeAPNICText(text)
	return c.classifyAPNICInetnumNormalized(text), c.classifyAPNICAccessNormalized(text)
}

func normalizeAPNICText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
