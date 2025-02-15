package dkui

import (
	"context"
	"database/sql"

	"dreamkast-weaver/internal/derrors"
	"dreamkast-weaver/internal/dkui/domain"
	"dreamkast-weaver/internal/dkui/repo"
	"dreamkast-weaver/internal/dkui/value"
	"dreamkast-weaver/internal/sqlhelper"
	"dreamkast-weaver/internal/stacktrace"

	"github.com/ServiceWeaver/weaver"
)

type Service interface {
	CreateViewEvent(ctx context.Context, profile Profile, req CreateViewEventRequest) error
	StampOnline(ctx context.Context, profile Profile, slotID value.SlotID) error
	StampOnSite(ctx context.Context, profile Profile, req StampRequest) error
	ViewingEvents(ctx context.Context, profile Profile) (*domain.ViewEvents, error)
	StampChallenges(ctx context.Context, profile Profile) (*domain.StampChallenges, error)
}

type Profile struct {
	weaver.AutoMarshal
	ID       value.ProfileID
	ConfName value.ConfName
}

type CreateViewEventRequest struct {
	weaver.AutoMarshal
	TrackID value.TrackID
	TalkID  value.TalkID
	SlotID  value.SlotID
}

type StampRequest struct {
	weaver.AutoMarshal
	TrackID value.TrackID
	TalkID  value.TalkID
	SlotID  value.SlotID
}

type ServiceImpl struct {
	weaver.Implements[Service]
	weaver.WithConfig[config]

	sh     *sqlhelper.SqlHelper
	domain domain.DkUiDomain
}

var _ Service = (*ServiceImpl)(nil)

type config struct {
	DBUser     string `toml:"db_user"`
	DBPassword string `toml:"db_password"`
	DBEndpoint string `toml:"db_endpoint"`
	DBPort     string `toml:"db_port"`
	DBName     string `toml:"db_name"`
}

func (c *config) SqlOption() *sqlhelper.SqlOption {
	return &sqlhelper.SqlOption{
		User:     c.DBUser,
		Password: c.DBPassword,
		Endpoint: c.DBEndpoint,
		Port:     c.DBPort,
		DbName:   c.DBName,
	}
}

func NewService(sh *sqlhelper.SqlHelper) Service {
	return &ServiceImpl{sh: sh}
}

func (s *ServiceImpl) Init(ctx context.Context) error {
	opt := s.Config().SqlOption()
	if err := opt.Validate(); err != nil {
		opt = sqlhelper.NewOptionFromEnv("dkui")
	}
	sh, err := sqlhelper.NewSqlHelper(opt)
	if err != nil {
		return err
	}
	s.sh = sh
	return nil
}

func (s *ServiceImpl) HandleError(msg string, err error) {
	if err != nil {
		if derrors.IsUserError(err) {
			s.Logger().With("errorType", "user-side").Info(msg, err)
		} else {
			s.Logger().With("stacktrace", stacktrace.Get(err)).Error(msg, err)
		}
	}
}

func (s *ServiceImpl) CreateViewEvent(ctx context.Context, profile Profile, req CreateViewEventRequest) (err error) {
	defer func() {
		s.HandleError("create viewEvent", err)
	}()

	r := repo.NewDkUiRepo(s.sh.DB())

	devents, err := r.ListViewEvents(ctx, profile.ConfName, profile.ID)
	if err != nil {
		return err
	}
	dstamps, err := r.GetTrailMapStamps(ctx, profile.ConfName, profile.ID)
	if err != nil {
		return err
	}

	ev, err := s.domain.CreateOnlineViewEvent(req.TrackID, req.TalkID, req.SlotID, dstamps, devents)
	if err != nil {
		return err
	}

	if err := s.sh.RunTX(ctx, func(ctx context.Context, tx *sql.Tx) error {
		r := repo.NewDkUiRepo(tx)
		if err := r.InsertViewEvents(ctx, profile.ConfName, profile.ID, ev); err != nil {
			return err
		}
		if err := r.UpsertTrailMapStamps(ctx, profile.ConfName, profile.ID, dstamps); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (v *ServiceImpl) ViewingEvents(ctx context.Context, profile Profile) (resp *domain.ViewEvents, err error) {
	defer func() {
		v.HandleError("get viewingEvents", err)
	}()

	r := repo.NewDkUiRepo(v.sh.DB())

	resp, err = r.ListViewEvents(ctx, profile.ConfName, profile.ID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (v *ServiceImpl) StampChallenges(ctx context.Context, profile Profile) (resp *domain.StampChallenges, err error) {
	defer func() {
		v.HandleError("get stampChallenges", err)
	}()

	r := repo.NewDkUiRepo(v.sh.DB())

	resp, err = r.GetTrailMapStamps(ctx, profile.ConfName, profile.ID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (v *ServiceImpl) StampOnline(ctx context.Context, profile Profile, slotID value.SlotID) (err error) {
	defer func() {
		v.HandleError("stamp from online", err)
	}()

	r := repo.NewDkUiRepo(v.sh.DB())

	dstamps, err := r.GetTrailMapStamps(ctx, profile.ConfName, profile.ID)
	if err != nil {
		return err
	}

	if err := v.domain.StampOnline(slotID, dstamps); err != nil {
		return err
	}

	if err := r.UpsertTrailMapStamps(ctx, profile.ConfName, profile.ID, dstamps); err != nil {
		return err
	}

	return nil
}

func (v *ServiceImpl) StampOnSite(ctx context.Context, profile Profile, req StampRequest) (err error) {
	defer func() {
		v.HandleError("stamp from onsite", err)
	}()

	r := repo.NewDkUiRepo(v.sh.DB())

	dstamps, err := r.GetTrailMapStamps(ctx, profile.ConfName, profile.ID)
	if err != nil {
		return err
	}

	ev, err := v.domain.StampOnSite(req.TrackID, req.TalkID, req.SlotID, dstamps)
	if err != nil {
		return err
	}

	if err := v.sh.RunTX(ctx, func(ctx context.Context, tx *sql.Tx) error {
		r := repo.NewDkUiRepo(tx)
		if err := r.InsertViewEvents(ctx, profile.ConfName, profile.ID, ev); err != nil {
			return err
		}
		if err := r.UpsertTrailMapStamps(ctx, profile.ConfName, profile.ID, dstamps); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
