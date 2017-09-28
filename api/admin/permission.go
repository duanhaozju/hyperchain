package jsonrpc

import (
	"fmt"
	"time"
)

// grantpermission grants the user with some permissions, permission must
// be specified with a module name(used to specify group) and a permission
// name. If user gave some non-existing permissions, returns to client.
func (admin *Administrator) grantpermission(username string, module string, permissions []string) ([]string, error) {
	// record non-existing permissions
	var invalidPms []string
	admin.logger.Debugf("grant permissions %v to %s module: %s", permissions, username, module)
	if _, err := admin.isUserExist(username, ""); err == ErrUserNotExist {
		return nil, err
	}
	//grant all permissions
	if toUpper(module) == "ALL" {
		admin.user_scope[username] = rootScopes()
		return nil, nil
	}
	// grant permissions one by one
	for _, pms := range permissions {
		if toUpper(pms) == "ALL" {
			pms = "all"
		}
		scope := convertToIntegers(module + "::" + pms)
		if scope == nil {
			invalidPms = append(invalidPms, module+"::"+pms)
			admin.logger.Noticef("Invalid permission name: %s", module+"::"+pms)
			continue
		}
		for _, permission := range scope {
			admin.user_scope[username][permission] = true
		}
	}

	return invalidPms, nil
}

// revokepermission revokes the user of some permissions, permission must
// be specified with a module name(used to specify group) and a permission
// name. If user gave some non-existing permissions, returns to client.
func (admin *Administrator) revokepermission(username string, module string, permissions []string) ([]string, error) {
	// record non-existing permissions
	var invalidPms []string
	admin.logger.Debugf("revoke permissions %v to %s", permissions, username)
	if _, err := admin.isUserExist(username, ""); err == ErrUserNotExist {
		return nil, err
	}
	// clear all permissions
	if toUpper(module) == "ALL" {
		admin.user_scope[username] = make(permissionSet)
		return nil, nil
	}
	// revoke permissions one by one
	for _, pms := range permissions {
		if toUpper(pms) == "ALL" {
			pms = "all"
		}
		scope := convertToIntegers(module + "::" + pms)
		if scope == nil {
			invalidPms = append(invalidPms, module+"::"+pms)
			admin.logger.Noticef("Invalid permission name: %s", module+"::"+pms)
			continue
		}
		for _, permission := range scope {
			delete(admin.user_scope[username], permission)
		}
	}

	return invalidPms, nil
}

// listpermission lists the permissions of the given user.
func (admin *Administrator) listpermission(username string) ([]int, error) {
	var permissions []int
	admin.logger.Debugf("list permissions of %s", username)
	if _, err := admin.isUserExist(username, ""); err == ErrUserNotExist {
		return nil, err
	}
	for scope := range admin.user_scope[username] {
		permissions = append(permissions, scope)
	}
	return permissions, nil
}

// checkPermission checks permission by username in input claims.
func (admin *Administrator) checkPermission(username, method string) (bool, error) {
	scope := convertToScope(method)
	if scope == -1 {
		return false, ErrPermission
	}

	if admin.isUserPermit(username, scope) {
		return true, nil
	} else {
		return false, ErrPermission
	}

}

// createUser creates a new account with the given username and password.
func (admin *Administrator) createuser(username, password, group string) error {
	groupPermission := getGroupPermission(group)
	if groupPermission == nil {
		return fmt.Errorf("Unrecoginzed group %s", group)
	}
	admin.valid_user[username] = password
	admin.user_scope[username] = groupPermission
	return nil
}

// alterUser alters an existed account with given username and password.
func (admin *Administrator) alteruser(username, password string) {
	admin.valid_user[username] = password
}

// delUser deletes an account.
func (admin *Administrator) deluser(username string) {
	delete(admin.valid_user, username)
	delete(admin.user_scope, username)
}

// checkOpTimeExpire checks if the given user's operation time has expired.
func (admin *Administrator) checkOpTimeExpire(username string) bool {
	return float64(time.Now().Unix()-admin.user_opTime[username]) > admin.expiration.Seconds()
}

// updateLastOperationTime updates the last operation time.
func (admin *Administrator) updateLastOperationTime(username string) {
	admin.user_opTime[username] = time.Now().Unix()
}

// isUserExist checks the username in valid_user, if username is not existed,
// returns ErrUserNotExist; if username is existed but password is not correct,
// returns ErrUnMatch; else returns true.
func (admin *Administrator) isUserExist(username, password string) (bool, error) {
	for k, v := range admin.valid_user {
		if k == username {
			if v == password {
				return true, nil
			} else {
				return false, ErrUnMatch
			}
		}
	}
	return false, ErrUserNotExist
}

// isUserPermit returns if the user has the permission with the given scope.
func (admin *Administrator) isUserPermit(username string, scope int) bool {
	for name, pSet := range admin.user_scope {
		if name == username {
			return pSet[scope]
		}
	}
	return false
}