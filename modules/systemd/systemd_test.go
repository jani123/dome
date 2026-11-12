package systemd_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/modules/systemd"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type runner struct{ mock.Mock }

func (r *runner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	a := r.Called(name, args)
	return a.Get(0).([]byte), a.Error(1)
}

func TestServiceObserveApply(t *testing.T) {
	r := &runner{}
	s := &systemd.Service{Name: "demo.service", Enabled: true, Running: true, Runner: r}
	require.NoError(t, s.Validate())
	r.On("Run", "systemctl", []string{"show", "--property=LoadState,ActiveState,UnitFileState", "--", "demo.service"}).Return([]byte("LoadState=loaded\nActiveState=active\nUnitFileState=enabled\n"), nil)
	v, err := s.Observe(context.Background())
	require.NoError(t, err)
	require.Equal(t, s.Desired(), v)
	for _, command := range []string{"enable", "start", "stop", "disable"} {
		r.On("Run", "systemctl", []string{command, "--", "demo.service"}).Return([]byte(nil), nil).Once()
	}
	require.NoError(t, s.Apply(context.Background(), resource.Update))
	require.NoError(t, s.Apply(context.Background(), resource.Delete))
	r.AssertExpectations(t)
}

func TestServiceRejectsUnsafeOrUnknownState(t *testing.T) {
	require.Error(t, (&systemd.Service{Name: "--evil.service"}).Validate())
	r := &runner{}
	r.On("Run", "systemctl", mock.Anything).Return([]byte("LoadState=not-found"), nil)
	_, err := (&systemd.Service{Name: "x.service", Runner: r}).Observe(context.Background())
	require.Error(t, err)
	bad := &runner{}
	sentinel := errors.New("systemctl failed")
	bad.On("Run", "systemctl", mock.Anything).Return([]byte(nil), sentinel)
	require.ErrorIs(t, (&systemd.Service{Name: "x.service", Runner: bad}).Apply(context.Background(), resource.Update), sentinel)
}
