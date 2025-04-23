package utils

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Types struct {
	Bool boolean
	Text text
	Time ti
	UUID uu
}

type boolean struct {}

func (b *boolean) ConvertToBool(val bool) pgtype.Bool {
	return pgtype.Bool{
		Bool: val,
		Valid: true,
	}
}

func (b *boolean) ConvertFromBool(val pgtype.Bool) bool {
	return val.Bool
}

type text struct {}

func (t *text) ConvertToText(val string) pgtype.Text {
	return pgtype.Text{
		String: val,
		Valid: true,
	}
}

func (t *text) ConvertFromText(val pgtype.Text) string {
	return val.String
}

type ti struct {}

func (t *ti) ConvertToTimestamp(val time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time: val,
		Valid: true,
	}
}

func (t *ti) ConvertFromTimestamp(val pgtype.Timestamptz) time.Time {
	return val.Time
}

type uu struct {}

func (u *uu) ConvertToUUID(val uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: val,
		Valid: true,
	}
}

func (u *uu) ConvertFromUUID(val pgtype.UUID) uuid.UUID {
	return val.Bytes
}