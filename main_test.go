package main

import "testing"

func TestListenAddress(t *testing.T) {
	for _, test := range []struct{ host, port, want string }{
		{"", "", ":3000"},
		{"127.0.0.1", "3000", "127.0.0.1:3000"},
		{"::1", "8080", "[::1]:8080"},
	} {
		t.Run(test.want, func(t *testing.T) {
			t.Setenv("HOST", test.host)
			t.Setenv("PORT", test.port)
			if got := listenAddress(); got != test.want {
				t.Fatalf("listenAddress() = %q, want %q", got, test.want)
			}
		})
	}
}
