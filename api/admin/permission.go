package jsonrpc

import (
	"fmt"
	"time"
)

// grantpermission grants the user with some permissions, permission must
// be specified with a module name(used to specify group) and a permission
// name. If user gave some non-existing permissions, returns to client.
func (adm *Administrator) grantpermission(username string, module string, permissions []string) ([]string, error) {
	// record non-existing permissions
	var invalidPms []string
	log.Debugf("grant permissions %v to %s module: %s", permissions, username, module)
	if _, err := isUserExist(username, ""); err == ErrUserNotExist {
		return nil, err
	}
	//grant all permissions
	if toUpper(module) == "ALL" {
		user_scope[username] = rootScopes()
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
			log.Noticef("Invalid permission name: %s", module+"::"+pms)
			continue
		}
		for _, permission := range scope {
			user_scope[username][permission] = true
		}
	}

	return invalidPms, nil
}

// revokepermission revokes the user of some permissions, permission must
// be specified with a module name(used to specify group) and a permission
// name. If user gave some non-existing permissions, returns to client.
func (adm *Administrator) revokepermission(username string, module string, permissions []string) ([]string, error) {
	// record non-existing permissions
	var invalidPms []string
	log.Debugf("revoke permissions %v to %s", permissions, username)
	if _, err := isUserExist(username, ""); err == ErrUserNotExist {
		return nil, err
	}
	// clear all permissions
	if toUpper(module) == "ALL" {
		user_scope[username] = make(permissionSet)
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
			log.Noticef("Invalid permission name: %s", module+"::"+pms)
			continue
		}
		for _, permission := range scope {
			delete(user_scope[username], permission)
		}
	}

	return invalidPms, nil
}

// listpermission lists the permissions of the given user.
func (adm *Administrator) listpermission(username string) ([]int, error) {
	var permissions []int
	log.Debugf("list permissions of %s", username)
	if _, err := isUserExist(username, ""); err == ErrUserNotExist {
		return nil, err
	}
	for scope := range user_scope[username] {
		permissions = append(permissions, scope)
	}
	return permissions, nil
}

// checkPermission checks permission by username in input claims.
func (adm *Administrator) checkPermission(username, method string) (bool, error) {
	scope := convertToScope(method)
	if scope == -1 {
		return false, ErrPermission
	}

	if isUserPermit(username, scope) {
		return true, nil
	} else {
		return false, ErrPermission
	}

}

// createUser creates a new account with the given username and password.
func (adm *Administrator) createuser(username, password, group string) error {
	groupPermission := getGroupPermission(group)
	if groupPermission == nil {
		return fmt.Errorf("Unrecoginzed group %s", group)
	}
	valid_user[username] = password
	user_scope[username] = groupPermission
	return nil
}

// alterUser alters an existed account with given username and password.
func (adm *Administrator) alteruser(username, password string) {
	valid_user[username] = password
}

// delUser deletes an account.
func (adm *Administrator) deluser(username string) {
	delete(valid_user, username)
	delete(user_scope, username)
}

// checkOpTimeExpire checks if the given user's operation time has expired.
func (adm *Administrator) checkOpTimeExpire(username string) bool {
	return float64(time.Now().Unix()-user_opTime[username]) > adm.Expiration.Seconds()
}