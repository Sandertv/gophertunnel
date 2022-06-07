package protocol

// DeviceOS is a device DeviceOS identifier. It holds a value of one of the constants below and may be found
// in packets such as the Login packet.
type DeviceOS int

const (
	DeviceAndroid DeviceOS = iota + 1
	DeviceIOS
	DeviceOSX
	DeviceFireOS
	DeviceGearVR
	DeviceHololens
	DeviceWin10
	DeviceWin32
	DeviceDedicated
	DeviceTVOS
	DeviceOrbis
	DeviceNX
	DeviceXBOX
	DeviceWP
	DeviceLinux
)
