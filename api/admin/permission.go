package jsonrpc

func grantpermission(username string, module string, permissions []string) ([]string, error) {
	// record non-existing permissions
	var invalidPms []string
	log.Debugf("grant permissions %v to %s module: %s", permissions, username, module)
	if _, err := IsUserExist(username, ""); err == ErrUserNotExist {
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

func revokepermission(username string, module string, permissions []string) ([]string, error) {
	// record non-existing permissions
	var invalidPms []string
	log.Debugf("revoke permissions %v to %s", permissions, username)
	if _, err := IsUserExist(username, ""); err == ErrUserNotExist {
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

func listpermission(username string) ([]int, error) {
	var permissions []int
	log.Debugf("list permissions of %s", username)
	if _, err := IsUserExist(username, ""); err == ErrUserNotExist {
		return nil, err
	}
	for scope := range user_scope[username] {
		permissions = append(permissions, scope)
	}
	return permissions, nil
}
