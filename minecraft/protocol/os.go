package protocol

// DeviceOS is a device DeviceOS identifier. It holds a value of one of the constants below and may be found
// in packets such as the Login packet.
type DeviceOS int

const (
	DeviceAndroid DeviceOS = iota + 1
	DeviceIOS
	DeviceOSX
	DeviceFireOS
	// Deprecated: DeviceGearVR is deprecated as of 1.21.50.
	DeviceGearVR
	DeviceHololens
	DeviceWin10
	DeviceWin32
	DeviceDedicated
	// Deprecated: DeviceTVOS is deprecated as of 1.20.10.
	DeviceTVOS
	DeviceOrbis // PlayStation
	DeviceNX
	DeviceXBOX
	// Deprecated: DeviceWP is deprecated as of 1.20.10.
	DeviceWP // Windows Phone
	DeviceLinux
)
