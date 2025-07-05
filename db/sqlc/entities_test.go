package db

//
// import (
// 	"context"
// 	"testing"
//
// 	"github.com/google/uuid"
// 	"github.com/jackc/pgx/v5/pgtype"
// 	"github.com/stretchr/testify/require"
// )
//
// func parentIDOrNil(id *uuid.UUID) [16]byte {
// 	var b [16]byte
// 	if id != nil {
// 		copy(b[:], id[:])
// 	}
// 	return b
// }
//
// func createRandomEntity(ctx context.Context, t *testing.T, q *Queries, parentID *uuid.UUID) Entity {
// 	uid := uuid.New()
//
// 	arg := CreateEntityParams{
// 		Uuid: uid,
// 		ParentID: UUIDPtr(uuid.New()),
// 		Name: "Test Entity",
// 		Code: nil,
// 		Type:          "department",
// 		IsActive:      true,
// 		Hidden:        false,
// 		AccrualMethod: true,
// 		FyStartMonth:  1,
// 		Address:       []byte(`{"city": "Nairobi"}`),
// 		Picture: nil,
// 		Settings: []byte(`{"currency": "KES"}`),
// 	}
//
// 	entity, err := q.CreateEntity(ctx, arg)
// 	require.NoError(t, err)
// 	require.Equal(t, arg.Name, entity.Name)
// 	require.Equal(t, arg.Code, entity.Code)
//
// 	return entity
// }
//
// func TestCreateAndGetEntity(t *testing.T) {
// 	setupTestDB(context.Background(), t)
//
// 	withTenantContext(t, 1, func(ctx context.Context, q *Queries) {
// 		entity := createRandomEntity(ctx, t, q, nil)
//
// 		found, err := q.GetEntity(ctx, entity.Uuid)
// 		require.NoError(t, err)
// 		require.Equal(t, entity.Uuid, found.Uuid)
// 	})}
// }
//
// func TestUpdateEntity(t *testing.T) {
// 	setupTestDB(context.Background(), t)
//
// 	withTenantContext(t, 1, func(q *Queries) {
// 		ctx := context.Background()
//
// 		entity := createRandomEntity(ctx, t, q, nil)
//
// 		newName := "Updated Entity"
// 		newCode := pgtype.Text{String: "UPDATED_" + entity.Uuid.String()[0:6], Valid: true}
// 		newType := "division"
//
// 		updated, err := q.UpdateEntity(ctx, UpdateEntityParams{
// 			Uuid:          entity.Uuid,
// 			Name:          newName,
// 			Code:          newCode,
// 			Type:          newType,
// 			IsActive:      false,
// 			Hidden:        false,
// 			AccrualMethod: false,
// 			FyStartMonth:  0,
// 			Address:       nil,
// 			Picture: pgtype.Text{
// 				String: "nil",
// 				Valid:  true,
// 			},
// 			Settings: nil,
// 		})
//
// 		require.NoError(t, err)
// 		require.Equal(t, newName, updated.Name)
// 		require.Equal(t, newCode.String, updated.Code.String)
// 		require.Equal(t, newType, updated.Type)
// 	})
// }
//
// func TestSoftDeleteEntity(t *testing.T) {
// 	setupTestDB(context.Background(), t)
//
// 	withTenantContext(t, 1, func(q *Queries) {
// 		ctx := context.Background()
//
// 		entity := createRandomEntity(ctx, t, q, nil)
//
// 		err := q.SoftDeleteEntity(ctx, entity.Uuid)
// 		require.NoError(t, err)
//
// 		_, err = q.GetEntity(ctx, entity.Uuid)
// 		require.Error(t, err)
// 	})
// }
//
// func TestListEntities(t *testing.T) {
// 	setupTestDB(context.Background(), t)
//
// 	withTenantContext(t, 1, func(q *Queries) {
// 		ctx := context.Background()
//
// 		_ = createRandomEntity(ctx, t, q, nil)
//
// 		list, err := q.ListEntities(ctx)
// 		require.NoError(t, err)
// 		require.NotEmpty(t, list)
// 	})
// }
