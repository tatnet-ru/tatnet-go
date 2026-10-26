package tatnet

// Preserve the published target-type constants. Adding the VM plan role enum
// makes oapi-codegen qualify these names to disambiguate the shared "vm" value.
const (
	App = LBTargetGroupSpecTargetTypeApp
	Ip  = LBTargetGroupSpecTargetTypeIp
	Vm  = LBTargetGroupSpecTargetTypeVm
)
