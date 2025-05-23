// autogenerated code
package gogen

////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////
//// Systems for 0 types
///

////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////
//// Entities num 0
///

var _ bool = _Entity_constraints(false)

func _Entity_constraints(v bool) bool {
	if !v {
		return true
	}

	return true
}

////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////
//// Components num 0
///

////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////
//// Queries num 0
///

var _ bool = _Query_constraints(false)

func _Query_constraints(v bool) bool {
	if v {
	}

	return true
}

////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////
//// Functions for 0 types
///

type EcsDebugInfo struct {
	EntitiesCount      int64
	EntitesCountByName map[string]int64
}

func MakeEcsDebugInfo() EcsDebugInfo {
	return EcsDebugInfo{
		EntitesCountByName: map[string]int64{},
	}
}

func (i EcsDebugInfo) Diff(prev EcsDebugInfo) EcsDebugInfo {
	r := MakeEcsDebugInfo()

	r.EntitiesCount = i.EntitiesCount - prev.EntitiesCount
	r.EntitesCountByName = i.EntitesCountByName
	for k, v := range prev.EntitesCountByName {
		r.EntitesCountByName[k] = i.EntitesCountByName[k] - v
	}

	return r
}

func GetEcsDebugInfo() EcsDebugInfo {
	info := MakeEcsDebugInfo()

	return info
}
