package device

// OS is a device OS identifier. It holds a value of one of the constants below and may be found in packets
// such as the Login packet.
type OS int

const (
	Android OS = iota + 1
	IOS
	OSX
	FireOS
	GearVR
	Hololens
	Win10
	Win32
	Dedicated
	TVOS
	Orbis
	NX
	XBOX
)
