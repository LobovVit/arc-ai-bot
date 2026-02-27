package guard

import (
	"regexp"
	"strings"
)

type Mode string

const (
	ModeOff Mode = "off"
	ModeOn  Mode = "on"
)

type Result struct {
	Removed   int
	Remaining int
}

var (
	// Минимальный набор сигнатур промпт-инъекций/опасных конструкций.
	// Можно расширять по мере нужды.
	injectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bignore\s+all\s+instructions\b`),
		regexp.MustCompile(`(?i)\bdisregard\s+previous\b`),
		regexp.MustCompile(`(?i)\byou\s+are\s+chatgpt\b`),
		regexp.MustCompile(`(?i)\b(system|developer)\s*:`),
		regexp.MustCompile(`(?i)\boutput\s*:`),
		regexp.MustCompile(`(?i)\bdo\s+not\s+follow\b`),
		regexp.MustCompile(`(?i)\bprompt\s+injection\b`),
	}

	// Маркеры секретов/паролей.
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bpassword\b`),
		regexp.MustCompile(`(?i)\bпарол[ья]\b`),
		regexp.MustCompile(`(?i)\broot\b`),
		regexp.MustCompile(`(?i)\bswordfish\b`),
		regexp.MustCompile(`(?i)\bapi[_ -]?key\b`),
		regexp.MustCompile(`(?i)\btoken\b`),
	}
)

// LooksMalicious returns true if text looks like prompt injection or secret payload.
func LooksMalicious(text string) bool {
	t := strings.TrimSpace(text)
	if t == "" {
		return false
	}
	for _, re := range injectionPatterns {
		if re.MatchString(t) {
			return true
		}
	}
	for _, re := range secretPatterns {
		if re.MatchString(t) {
			return true
		}
	}
	return false
}

// Sanitize removes obvious instruction-like lines from context.
// Use if you prefer to keep some chunk text instead of dropping it.
func Sanitize(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		l := strings.TrimSpace(ln)
		if l == "" {
			out = append(out, ln)
			continue
		}
		low := strings.ToLower(l)
		if strings.Contains(low, "ignore all instructions") ||
			strings.HasPrefix(low, "system:") ||
			strings.HasPrefix(low, "developer:") ||
			strings.HasPrefix(low, "output:") {
			continue
		}
		out = append(out, ln)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// FilterChunks drops chunks that look malicious.
func FilterChunks[T any](chunks []T, getText func(T) string) ([]T, Result) {
	out := make([]T, 0, len(chunks))
	removed := 0
	for _, c := range chunks {
		if LooksMalicious(getText(c)) {
			removed++
			continue
		}
		out = append(out, c)
	}
	return out, Result{Removed: removed, Remaining: len(out)}
}

// LooksLikeSecretQuestion blocks requests that try to extract secrets.
func LooksLikeSecretQuestion(q string) bool {
	for _, re := range secretPatterns {
		if re.MatchString(q) {
			return true
		}
	}
	// Добавим типовые формулировки
	low := strings.ToLower(q)
	if strings.Contains(low, "суперпароль") ||
		strings.Contains(low, "пароль root") ||
		strings.Contains(low, "root-пользователя") ||
		strings.Contains(low, "назови пароль") ||
		strings.Contains(low, "скажи пароль") {
		return true
	}
	return false
}
