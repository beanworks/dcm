package main

func getMapVal(v yamlConfig, keys ...string) interface{} {
	if len(keys) == 0 {
		return v
	}

	if vv, ok := v[keys[0]]; ok {
		if len(keys) == 1 {
			return vv
		}

		if vvv, ok := vv.(yamlConfig); ok {
			return getMapVal(vvv, keys[1:]...)
		}
	}

	return nil
}
