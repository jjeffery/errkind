package errkind

import (
	"testing"

	"github.com/jjeffery/errors"
)

func TestCode(t *testing.T) {
	tests := []struct {
		err      error
		codes    []string
		want     bool
		wantCode string
	}{
		{
			err:      nil,
			codes:    []string{"A", "B", "C"},
			want:     false,
			wantCode: "",
		},
		{
			err:      PublicWithCode("test error", 0, "CODE"),
			codes:    []string{"A", "B", "C"},
			want:     false,
			wantCode: "CODE",
		},
		{
			err:      PublicWithCode("test error", 0, "CODE").With("a", "b"),
			codes:    []string{"A", "B", "C"},
			want:     false,
			wantCode: "CODE",
		},
		{
			err:      PublicWithCode("test error", 0, "").With("a", "b"),
			codes:    []string{"A", "B", "C"},
			want:     false,
			wantCode: "",
		},
		{
			err:      PublicWithCode("test error", 0, "CODE"),
			codes:    []string{"A", "B", "CODE"},
			want:     true,
			wantCode: "CODE",
		},
		{
			err:      PublicWithCode("test error", 0, "CODE").With("a", "b"),
			codes:    []string{"A", "B", "CODE"},
			want:     true,
			wantCode: "CODE",
		},
		{
			err:      errors.New("test error").With("a", "b"),
			codes:    []string{"A", "B", "C"},
			want:     false,
			wantCode: "",
		},
	}
	for i, tt := range tests {
		if got, want := HasCode(tt.err, tt.codes...), tt.want; want != got {
			t.Errorf("%d: want=%v, got=%v", i, want, got)
		}
		if got, want := Code(tt.err), tt.wantCode; want != got {
			t.Errorf("%d: want=%v, got=%v", i, want, got)
		}
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		err        error
		statuses   []int
		want       bool
		wantStatus int
	}{
		{
			err:        nil,
			statuses:   []int{500},
			want:       false,
			wantStatus: 0,
		},
		{
			err:        Public("test error", 501),
			statuses:   []int{400, 401, 402},
			want:       false,
			wantStatus: 501,
		},
		{
			err:        PublicWithCode("test error", 501, "CODE").With("a", "b"),
			statuses:   []int{400, 401, 402},
			want:       false,
			wantStatus: 501,
		},
		{
			err:        Public("test error", 501),
			statuses:   []int{500, 501},
			want:       true,
			wantStatus: 501,
		},
		{
			err:        PublicWithCode("test error", 400, "CODE").With("a", "b"),
			statuses:   []int{400},
			want:       true,
			wantStatus: 400,
		},
		{
			err:        testingStatusError(501),
			statuses:   []int{400, 401, 402},
			want:       false,
			wantStatus: 501,
		},
		{
			err:        testingStatusError(402),
			statuses:   []int{400, 401, 402},
			want:       true,
			wantStatus: 402,
		},
		{
			err:        errors.New("no status"),
			statuses:   []int{400},
			want:       false,
			wantStatus: 0,
		},
	}
	for i, tt := range tests {
		if got, want := HasStatus(tt.err, tt.statuses...), tt.want; want != got {
			t.Errorf("%d: want=%v, got=%v", i, want, got)
		}
		if got, want := Status(tt.err), tt.wantStatus; want != got {
			t.Errorf("%d: want=%v, got=%v", i, want, got)
		}
	}
}

type testingStatusError int

func (err testingStatusError) Error() string {
	return "testing status error"
}

func (err testingStatusError) Status() int {
	return int(err)
}

func TestTemporary(t *testing.T) {
	tests := []struct {
		err       error
		want      bool
		wantError string
	}{
		{
			err:  nil,
			want: false,
		},
		{
			err:       Temporary("temp"),
			want:      true,
			wantError: "temp",
		},
		{
			err:       errors.Wrap(Temporary("temp"), "wrapped").With("a", "b"),
			want:      true,
			wantError: "wrapped a=b: temp",
		},
		{
			err:       errors.New("not temporary"),
			want:      false,
			wantError: "not temporary",
		},
		{
			err:       errors.Wrap(errors.New("not temporary"), "wrapped"),
			want:      false,
			wantError: "wrapped: not temporary",
		},
	}
	for i, tt := range tests {
		if got, want := IsTemporary(tt.err), tt.want; got != want {
			t.Errorf("%d: want=%v, got=%v", i, want, got)
		}
		if tt.err != nil {
			if got, want := tt.err.Error(), tt.wantError; got != want {
				t.Errorf("%d: want=%v, got=%v", i, want, got)
			}
		}
	}
}

func TestPublic(t *testing.T) {
	tests := []struct {
		err        error
		want       bool
		wantError  string
		wantStatus int
		wantCode   string
	}{
		{
			err:        nil,
			want:       false,
			wantStatus: 0,
		},
		{
			err:        Public("public", 400),
			want:       true,
			wantError:  "public",
			wantStatus: 400,
		},
		{
			err:        errors.Wrap(Public("public", 401), "wrapped").With("a", "b"),
			want:       false,
			wantError:  "wrapped a=b: public",
			wantStatus: 401,
		},
		{
			err:        PublicWithCode("public", 400, "XXX"),
			want:       true,
			wantError:  "public",
			wantStatus: 400,
			wantCode:   "XXX",
		},
		{
			err:        errors.Wrap(PublicWithCode("public", 401, "YYY"), "wrapped").With("a", "b"),
			want:       false,
			wantError:  "wrapped a=b: public",
			wantStatus: 401,
			wantCode:   "YYY",
		},
		{
			err:        errors.New("not public"),
			want:       false,
			wantError:  "not public",
			wantStatus: 0,
		},
		{
			err:        errors.Wrap(errors.New("not public"), "wrapped"),
			want:       false,
			wantError:  "wrapped: not public",
			wantStatus: 0,
		},
	}
	for i, tt := range tests {
		if got, want := IsPublic(tt.err), tt.want; got != want {
			t.Errorf("%d: want=%v, got=%v", i, want, got)
		}
		if got, want := Status(tt.err), tt.wantStatus; got != want {
			t.Errorf("%d: want=%v, got=%v", i, want, got)
		}
		if got, want := Code(tt.err), tt.wantCode; got != want {
			t.Errorf("%d: want=%v, got=%v", i, want, got)
		}
		if tt.err != nil {
			if got, want := tt.err.Error(), tt.wantError; got != want {
				t.Errorf("%d: want=%v, got=%v", i, want, got)
			}

		}
	}
}
