package apnicorg

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"strings"
)

// Parse returns APNIC organisation handles mapped to their structured org-name.
func Parse(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	out := map[string]string{}
	fields := map[string][]string{}
	last := ""
	finish := func() {
		if h, name := first(fields["organisation"]), first(fields["org-name"]); h != "" && name != "" {
			out[h] = name
		}
		fields, last = map[string][]string{}, ""
	}
	s := bufio.NewScanner(z)
	s.Buffer(make([]byte, 64*1024), 1024*1024)
	for s.Scan() {
		line := strings.TrimRight(s.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			finish()
			continue
		}
		if strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}
		if (line[0] == ' ' || line[0] == '\t' || line[0] == '+') && last != "" {
			v := fields[last]
			v[len(v)-1] = strings.TrimSpace(v[len(v)-1] + " " + strings.TrimSpace(strings.TrimPrefix(line, "+")))
			fields[last] = v
			continue
		}
		colon := strings.IndexByte(line, ':')
		if colon <= 0 {
			return nil, fmt.Errorf("%s: malformed RPSL line", path)
		}
		last = strings.ToLower(strings.TrimSpace(line[:colon]))
		fields[last] = append(fields[last], strings.TrimSpace(line[colon+1:]))
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	finish()
	if len(out) == 0 {
		return nil, fmt.Errorf("%s contains no organisation records", path)
	}
	return out, nil
}

func first(v []string) string {
	if len(v) == 0 {
		return ""
	}
	return v[0]
}
