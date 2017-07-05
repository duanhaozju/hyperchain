package jsonrpc

func grantpermission(username string, permissions []string) ([]string, error) {
	// record non-existing permissions
	var invalidPms []string
	log.Debugf("grant permissions %v to %s", permissions, username)
	if _, err:= IsUserExist(username, ""); err == ErrUserNotExist {
		return nil, err
	}
	// grant permissions one by one
	for _, pms := range permissions {
		scope := convertToScope(pms)
		if scope == -1 {
			invalidPms = append(invalidPms, pms)
			log.Noticef("Invalid permission name: %v", pms)
			continue
		}
		user_scope[username][scope] = true
	}

	return invalidPms, nil
}

func revokepermission(username string, permissions []string) ([]string, error) {
	// record non-existing permissions
	var invalidPms []string
	log.Debugf("revoke permissions %v to %s", permissions, username)
	if _, err:= IsUserExist(username, ""); err == ErrUserNotExist {
		return nil, err
	}
	// clear all permissions
	if toUpper(permissions[0]) == "ALL" {
		user_scope[username] = make(permissionSet)
		return nil, nil
	}
	// revoke permissions one by one
	for _, pms := range permissions {
		scope := convertToScope(pms)
		if scope == -1 {
			invalidPms = append(invalidPms, pms)
			log.Noticef("Invalid permission name: %v", pms)
			continue
		}
		delete(user_scope[username], scope)
	}

	return invalidPms, nil
}

func listpermission(username string) ([]int, error) {
	var permissions []int
	log.Debugf("list permissions of %s", username)
	if _, err:= IsUserExist(username, ""); err == ErrUserNotExist {
		return nil, err
	}
	for scope := range user_scope[username] {
		permissions = append(permissions, scope)
	}
	return permissions, nil
}