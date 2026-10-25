package config

import "testing"

func TestHasPermission(t *testing.T) {
	t.Run("AdminHasEveryPermission", func(t *testing.T) {
		if len(allPermissions) == 0 {
			t.Fatal("allPermissions is empty; init() did not run")
		}

		for _, perm := range allPermissions {
			if !HasPermission(RoleAdmin, perm) {
				t.Errorf("expected RoleAdmin to have permission %q", perm)
			}
		}
	})

	t.Run("UserGrantedPermission", func(t *testing.T) {
		if !HasPermission(RoleUser, PermCreateFarm) {
			t.Error("expected RoleUser to have PermCreateFarm")
		}
	})

	t.Run("GuestReadOnly", func(t *testing.T) {
		readPerms := []Permission{
			PermGetOneFarm,
			PermGetManyFarms,
			PermGetOnePost,
			PermGetManyPosts,
			PermGetOneProduct,
			PermGetManyProducts,
			PermGetLocationByFarmID,
		}
		for _, perm := range readPerms {
			if !HasPermission(RoleGuest, perm) {
				t.Errorf("expected RoleGuest to have read permission %q", perm)
			}
		}
	})

	t.Run("GuestDeniedWrite", func(t *testing.T) {
		writePerms := []Permission{
			PermCreateFarm,
			PermUpdateFarm,
			PermDeleteFarm,
			PermCreatePost,
			PermDeleteTransaction,
		}
		for _, perm := range writePerms {
			if HasPermission(RoleGuest, perm) {
				t.Errorf("expected RoleGuest to NOT have write permission %q", perm)
			}
		}
	})

	t.Run("UnknownRoleHasNoPermissions", func(t *testing.T) {
		if HasPermission(Role("BOGUS"), PermGetOneFarm) {
			t.Error("expected unknown role to have no permissions")
		}
	})

	t.Run("UnknownPermissionAlwaysDenied", func(t *testing.T) {
		if HasPermission(RoleAdmin, Permission("resource:bogus_action")) {
			t.Error("expected unknown permission to be denied even for RoleAdmin")
		}
	})
}
