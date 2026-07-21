package operatorconfig

import (
	"encoding/json"
	"fmt"
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
}

type configFile struct {
	Operators                      map[string]operator `json:"operators"`
	ExcludeDescriptionRules        []descriptionRule   `json:"exclude_description_rules"`
	ExcludeAPNICInetnumRules       []descriptionRule   `json:"exclude_apnic_inetnum_rules"`
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
}

type inclusion struct {
	operator string
	reason   string
}

type Classifier struct {
	rules               []rule
	included            map[string]inclusion
	excluded            map[string]string
	exclusionPatterns   []exclusionRule
	apnicPatterns       []exclusionRule
	legalEntityPatterns []*regexp.Regexp
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
		re, err := regexp.Compile("(?i)(?:" + rule.Pattern + ")")
		if err != nil {
			return nil, fmt.Errorf("APNIC inetnum exclusion pattern %q: %w", rule.Pattern, err)
		}
		c.apnicPatterns = append(c.apnicPatterns, exclusionRule{pattern: re, source: rule.Pattern, reason: rule.Reason})
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
	return c, nil
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
	text = strings.Join(strings.Fields(text), " ")
	for _, rule := range c.apnicPatterns {
		if rule.pattern.MatchString(text) {
			return PrefixResult{Excluded: true, Reason: rule.reason, MatchedBy: "exclude_apnic_inetnum_rules: " + rule.source}
		}
	}
	return PrefixResult{}
}
