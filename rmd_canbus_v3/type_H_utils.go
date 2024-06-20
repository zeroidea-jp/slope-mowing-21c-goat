package rmd_canbus_v3

const MOTOR_ID_1 = MotorID(0x01)
const MOTOR_ID_2 = MotorID(0x02)
const MOTOR_ID_3 = MotorID(0x03)
const MOTOR_ID_4 = MotorID(0x04)

type motions struct {
	current motion
	last    motion
}

type motion uint8

// mapt to safly access to motion correspoinding to driveMode
type motionMap map[driveMode]motion

type modes struct {
	current driveMode
	last    driveMode
}

type driveMode uint8

const emergecy = driveMode(0)
const standby = driveMode(1)
const auto = driveMode(2)

type switchState struct {
	isMainSwitchOn       isSwitchOn
	isAutoSwtichOn       isSwitchOn
	isEmergencyOn        isSwitchOn
	isWindingSwitch_1_On isSwitchOn
	isWindingSwitch_2_On isSwitchOn
	isWindingSwitch_3_On isSwitchOn
	isWindingSwitch_4_On isSwitchOn
}

type isSwitchOn bool

type SensorState struct {
	dummySensor int64
}
