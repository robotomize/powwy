package hashcash

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDefault(t *testing.T) {
	t.Parallel()

	t.Run("dump_test", func(t *testing.T) {
		t.Parallel()

		header, err := Default("127.0.0.1", 4, time.Now().Add(10*time.Minute).Unix())
		if err != nil {
			t.Fatalf("oops something wrong: %v", err)
		}

		tokens := strings.Split(header.String(), ":")

		if diff := cmp.Diff(7, len(tokens)); diff != "" {
			t.Fatalf("mismatch (-want, +got):\n%s", diff)
		}

		if diff := cmp.Diff("1", tokens[0]); diff != "" {
			t.Fatalf("mismatch (-want, +got):\n%s", diff)
		}

		if diff := cmp.Diff(AlgSHA256, tokens[4]); diff != "" {
			t.Fatalf("mismatch (-want, +got):\n%s", diff)
		}
	})
}

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		isErr    bool
		header   string
		expected Header
	}{
		{
			name:   "test_set_sha256_counter_0",
			header: "1:20:1665396610:localhost:sha-256:vZOxuoIgixP+hw==:AAAAAAAAAAA=",
			expected: Header{
				Version:   1,
				Difficult: 20,
				Subject:   "localhost",
				ExpiredAt: time.Date(2022, 10, 10, 10, 10, 10, 0, time.UTC).Unix(),
				Alg:       AlgSHA256,
				Nonce:     "vZOxuoIgixP+hw==",
				Counter:   0,
			},
		},
		{
			name:   "test_set_sha512_counter_5",
			header: "1:20:1665396610:localhost:sha-512:hVscDCMZcS1WYg==:BQAAAAAAAAA=",
			expected: Header{
				Version:   1,
				Difficult: 20,
				Subject:   "localhost",
				ExpiredAt: time.Date(2022, 10, 10, 10, 10, 10, 0, time.UTC).Unix(),
				Alg:       AlgSHA512,
				Nonce:     "hVscDCMZcS1WYg==",
				Counter:   5,
			},
		},
		{
			name:   "test_set_sha512_counter_5",
			header: "1:665396610:localhost:sha-512:hVscDCMZcS1WYg==:BQAAAAAAAAA=",
			isErr:  true,
			expected: Header{
				Version:   1,
				Difficult: 20,
				Subject:   "localhost",
				ExpiredAt: time.Date(2022, 10, 10, 10, 10, 10, 0, time.UTC).Unix(),
				Alg:       AlgSHA512,
				Nonce:     "hVscDCMZcS1WYg==",
				Counter:   5,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			header, err := Parse(tc.header)
			if err != nil {
				if !tc.isErr {
					t.Error(err)

					return
				}

				return
			}

			if diff := cmp.Diff(tc.expected, header); diff != "" {
				t.Fatalf("mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestCompute(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		header        string
		maxIterations int
		expected      string
	}{
		{
			name:          "test_calculate_set_sha256",
			header:        "1:5:1665396610:localhost:sha-256:vZOxuoIgixP+hw==:AAAAAAAAAAA=",
			maxIterations: 1 << 22,
			expected:      "0000036404f2d2f2d287320abf84fae7b1cbb48ee4d98e6ea8760596f6e07992",
		},
		{
			name:          "test_calculate_set_sha512",
			header:        "1:5:1665396610:localhost:sha-512:vZOxuoIgixP+hw==:AAAAAAAAAAA=",
			maxIterations: 1 << 22,
			expected:      "00000e738acbb0e365a15673af3b5d1d4149b8fcce8cc23eb68da76ee722ec06fd74acc2b3ca973160a7ac2953f6a78446632867a2543cb01698b661addd9258",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			header, err := Parse(tc.header)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			calculated, err := Compute(context.Background(), header, tc.maxIterations)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expected, calculated.Hash()); diff != "" {
				t.Fatalf("mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestComputeWithPool(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		header        string
		maxIterations int
		expected      string
	}{
		{
			name:          "test_calculate_set_sha256",
			header:        "1:5:1665396610:localhost:sha-256:vZOxuoIgixP+hw==:AAAAAAAAAAA=",
			maxIterations: 1 << 22,
			expected:      "0000036404f2d2f2d287320abf84fae7b1cbb48ee4d98e6ea8760596f6e07992",
		},
		{
			name:          "test_calculate_set_sha512",
			header:        "1:5:1665396610:localhost:sha-512:vZOxuoIgixP+hw==:AAAAAAAAAAA=",
			maxIterations: 1 << 26,
			expected:      "00000b57794d80276276d3f8fb65f57f4272eb0eb20fff2f04a3b3d0a22029093981e026c0e9db3a152bf6bf8bcb56a0f1ebcbe0a83676c1cdd065e2b9554b1a",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			header, err := Parse(tc.header)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			info, err := ComputeWithPool(
				ctx, header, WithWorkerNum(4), WithPoolMaxIterations(tc.maxIterations),
			)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expected, info.Header.Hash()); diff != "" {
				t.Fatalf("mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}
