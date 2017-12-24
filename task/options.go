package task

import "bufio"

// WithArgs specifies a slice of raw arguments as source input to a Factory.
// Behavior falls back to reading values from the Scanner if the supplied argument slice is empty.
func WithArgs(args []string) func(*Config) {
	return func(cfg *Config) {
		cfg.args = args
	}
}

// WithScanner overrides the default Scanner used for supplying values to a Factory.
// The default scanner tokenizes lines of text read from STDIN.
func WithScanner(s *bufio.Scanner) func(*Config) {
	return func(cfg *Config) {
		cfg.scanner = s
	}
}

// WithReporter overrides the reporting function applied to completed task results.
func WithReporter(fn func(completed Task)) func(*Config) {
	return func(cfg *Config) {
		cfg.reportFunc = fn
	}
}
