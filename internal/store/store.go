package store

import (
	"context"
	"database/sql"
	"time"
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetOne(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, *PaginatedFeedQuery) ([]PostWithMetadata, error)
	}
	Users interface {
		Create(context.Context, *sql.Tx, *User) error
		GetOne(context.Context, int64) (*User, error)
		GetOneByEmail(context.Context, string) (*User, error)
		FollowUser(context.Context, int64, int64) error
		UnfollowUser(context.Context, int64, int64) error
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		Activate(context.Context, string) error
		RollBackNewUser(context.Context, int64, string) error
	}
	Roles interface {
		// Create(context.Context, *sql.Tx, *User) error
		GetOneByName(context.Context, string) (*Role, error)
		GetAllAboveLevel(context.Context, int64) ([]Role, error)
	}
	Comments interface {
		Create(context.Context, *Comment) error
		GetPostWithCommentsByID(context.Context, int64) ([]Comment, error)
	}

	Students interface {
		RegisterAndInvite(context.Context, *Student, string, time.Duration) error
		RollBackNewStudent(context.Context, int64, string) error
		Activate(context.Context, string) error
		GetOneByEmail(context.Context, string) (*Student, error)
		GetOneByID(context.Context, int64) (*Student, error)

		GetStudentPersonalByID(context.Context, int64) (*StudentPersonal, error)
		CreateStudentPersonal(context.Context, StudentPersonal, int64) error
		UpdateStudentPersonal(context.Context, StudentPersonal, int64) error

		GetStudentInstitutionByID(context.Context, int64) (*StudentInstitution, error)
		CreateStudentInstitution(context.Context, StudentInstitution, int64) error
		UpdateStudentInstitution(context.Context, StudentInstitution, int64) error

		GetStudentSponsorByID(context.Context, int64) (*StudentSponsor, error)
		CreateStudentSponsor(context.Context, StudentSponsor, int64) error
		UpdateStudentSponsor(context.Context, StudentSponsor, int64) error

		GetStudentEmergencyByID(context.Context, int64) (*StudentEmergency, error)
		CreateStudentEmergency(context.Context, StudentEmergency, int64) error
		UpdateStudentEmergency(context.Context, StudentEmergency, int64) error

		GetStudentGuardiansByID(context.Context, int64) ([]StudentGuardian, error)
		CreateStudentGuardian(context.Context, StudentGuardian, int64) error
		UpdateStudentGuardian(context.Context, StudentGuardian, int64) error
		DeleteStudentGuardian(context.Context, int64, int64) error

		CreateStudentApplication(context.Context, int64, int64) error
		WithdrawStudentApplication(context.Context, int64, int64) error
		GetStudentApplications(context.Context, int64) ([]BursaryWithMetadata, error)
	}

	Admins interface {
		RegisterAndInvite(context.Context, *Admin, string, time.Duration) error
		RollBackNewAdmin(context.Context, int64, string) error
		Activate(context.Context, string) error
		GetOneByEmail(context.Context, string) (*Admin, error)
		GetOneByID(context.Context, int64) (*Admin, error)
		GetAdminUsers(context.Context, *PaginatedAdminUserQuery) ([]Admin, error)
		GetRoles(context.Context) ([]Role, error)
	}

	Bursaries interface {
		GetBursariesAndCount(context.Context, *PaginatedFeedQuery) (*BursariesWithMetadata, error)
		GetBursaries(context.Context, *sql.Tx, *PaginatedFeedQuery) ([]Bursary, error)
		GetBursaryApplications(context.Context, int64, int64) ([]BursaryWithMetadata, error)
		// GetBursaryByID(context.Context, int64) (*Bursary, error)
		GetBursariesCount(context.Context, *sql.Tx, *PaginatedFeedQuery) (int64, error)
		CreateBursary(context.Context, Bursary) error
		UpdateBursary(context.Context, Bursary) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Users:     &UserStore{db},
		Comments:  &CommentStore{db},
		Roles:     &RolesStore{db},
		Students:  &StudentsStore{db},
		Admins:    &AdminsStore{db},
		Bursaries: &BursariesStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
