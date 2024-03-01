package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

const MOTOR_ID_1 = MotorID(0x01)
const MOTOR_ID_2 = MotorID(0x02)

const FLAG_SINGLE_MOTOR MotorTargetFlag = 0x140

// const CURRENT_CONTRL_VAL

// const FLAG_MULTI_MOTOR MotorTargetFlag = 0x280

type MotorTargetFlag uint32

type MotorModelParameters struct {
}

type MotorID uint8
type Model string

// type MotorModels map[Model]struct {
// }

type Orders map[MotorID]*OrderData

type OrderData struct {
	canFrame can.Frame
}

func send_command_to_each_motors(ctx context.Context, pTx *socketcan.Transmitter, motor_ids []MotorID, s_data []can.Data) error {
	for i, motor_id := range motor_ids {

		err := send_command(ctx, pTx, motor_id, s_data[i], FLAG_SINGLE_MOTOR)
		if err != nil {
			log.Println("error when sending command")
			log.Println(err)
			return err
		}
	}

	return nil
}

func send_command(ctx context.Context, pTx *socketcan.Transmitter, motor_id MotorID, data can.Data, flag MotorTargetFlag) error {
	data_frame := can.Frame{
		ID:         uint32(motor_id) | uint32(flag),
		Length:     8,
		Data:       data,
		IsRemote:   false,
		IsExtended: false,
	}

	err := pTx.TransmitFrame(ctx, data_frame)
	if err != nil {
		log.Println("Error when transmitting Frame (1)")
	}
	return err

}

func receive_reply(ctx context.Context, pConn net.Conn, pWg *sync.WaitGroup) {
	pRecv := socketcan.NewReceiver(pConn)
	defer pRecv.Close()
LOOP:
	for pRecv.Receive() {
		frame := pRecv.Frame()
		fmt.Println(frame.String())
		select {
		case <-ctx.Done():
			log.Println("ctx.Done() proceeded.")
			break LOOP
		default:
		}
	}
	pWg.Done()
}

type current_in_deci_A_int16 int16
type angle_in_deci_degree_int16 int16
type angle_in_deci_degree_int32 int32

type shaft_angle_speed_in_deg_per_sec_int16 int16

func (deci_A current_in_deci_A_int16) to_can_data_alley() [2]uint8 {
	return (int16_to_uint8_alley(int16(deci_A)))
	// return integer_to_alley_uint8_x_2(deci_A)
}

func (deci_deg angle_in_deci_degree_int16) to_can_data_alley() [2]uint8 {
	return (int16_to_uint8_alley(int16(deci_deg)))
	// return integer_to_alley_uint8_x_2(deci_deg)
}

func (deci_deg angle_in_deci_degree_int32) to_can_data_alley() [4]uint8 {
	return (int32_to_uint8_alley(int32(deci_deg)))
	// return integer_to_alley_uint8_x_4(deci_deg)
}

func (deg_per_sec shaft_angle_speed_in_deg_per_sec_int16) to_can_data_alley() [2]uint8 {
	return (int16_to_uint8_alley(int16(deg_per_sec)))
	// return integer_to_alley_uint8_x_2(deg_per_sec)
}

func int16_to_uint8_alley(int16_num int16) [2]uint8 {
	return [2]uint8{
		uint8(int16_num & 0xFF),
		uint8(int16_num >> 8 & 0xFF),
	}
}

func int32_to_uint8_alley(int32_num int32) [4]uint8 {
	return [4]uint8{
		uint8(int32_num & 0xFF),
		uint8(int32_num >> 8 & 0xFF),
		uint8(int32_num >> 16 & 0xFF),
		uint8(int32_num >> 24 & 0xFF),
	}
}

func integer_to_uint8_alley(must_be_integer interface{}) [4]uint8 {
	uint8_alley_for_can_data := [4]uint8{0, 0, 0, 0}
	if integer, ok := must_be_integer.(int32); ok {
		for i, _ := range uint8_alley_for_can_data {
			uint8_alley_for_can_data[i] = uint8(integer >> (8 * i) & 0xFF)
		}
	} else if integer, ok := must_be_integer.(int16); ok {
		uint8_alley_for_can_data[0] = uint8(integer & 0xFF)
		uint8_alley_for_can_data[0] = uint8(integer >> 8 & 0xFF)
	}
	return uint8_alley_for_can_data
}

func main() {
	motor_ids := []MotorID{MOTOR_ID_1, MOTOR_ID_2}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	var wg sync.WaitGroup
	defer wg.Wait()

	pConn, err := socketcan.DialContext(ctx, "can", "can0")
	if err != nil {
		log.Println("Error when Dialialing with Context.")
		log.Fatal(err)
	}
	defer pConn.Close()

	pTx := socketcan.NewTransmitter(pConn)
	defer pTx.Close()

	wg.Add(1)
	go receive_reply(ctx, pConn, &wg)

	deci_A_cw := current_in_deci_A_int16(15)
	a_deci_A_cw := deci_A_cw.to_can_data_alley()
	deci_A_ccw := current_in_deci_A_int16(-15)
	a_deci_A_ccw := deci_A_ccw.to_can_data_alley()

	data_to_torque_closed_loop_control_cw := can.Data{0xA1, 0x00, 0x00, 0x00, a_deci_A_cw[0], a_deci_A_cw[1], 0x00, 0x00}
	data_to_torque_closed_loop_control_ccw := can.Data{0xA1, 0x00, 0x00, 0x00, a_deci_A_ccw[0], a_deci_A_ccw[1], 0x00, 0x00}

	// s_data_to_torque_closed_loop_control := []can.Data{
	// 	{0xA1, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00}, // Torque Control
	// 	{0xA1, 0x00, 0x00, 0x00, 0x13, 0x00, 0x00, 0x00}, // Torque Control
	// }

	data_to_read_multi_turn_encoder_position_data := can.Data{0x60, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	// data_to_set_multi_turn_position_zero_to_rom_RESTART_NEEDED := can.Data{0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00} // not woriking??

	shaft_angle_speed := shaft_angle_speed_in_deg_per_sec_int16(360)
	a_speed := shaft_angle_speed.to_can_data_alley()

	delta_deci_degree := angle_in_deci_degree_int32(0 * 100)
	a_delt_c_deg := delta_deci_degree.to_can_data_alley()
	data_to_control_absolute_position_in_closed_loop_0 := can.Data{0xA4, 0x00, a_speed[0], a_speed[1], a_delt_c_deg[0], a_delt_c_deg[1], a_delt_c_deg[2], a_delt_c_deg[3]} // works

	delta_deci_degree = angle_in_deci_degree_int32(360 * 100)
	a_delt_c_deg = delta_deci_degree.to_can_data_alley()
	data_to_control_absolute_position_in_closed_loop_360 := can.Data{0xA4, 0x00, a_speed[0], a_speed[1], a_delt_c_deg[0], a_delt_c_deg[1], a_delt_c_deg[2], a_delt_c_deg[3]} // works

	data_to_shutdown_motor := can.Data{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	// // Seems NOT to work for RMD X4
	// data_to_control_incremental_position_in_closed_loop := can.Data{0xA8, 0x00, 0x00, 0x02, s_delt_c_deg[0], s_delt_c_deg[1], s_delt_c_deg[2], s_delt_c_deg[3]}
	// data_to_control_incremental_position_in_closed_loop := can.Data{0xA8, 0x00, 0x00, 0x02, 0x0, 0x00, 0x00, 0x00} // does not stop
	// data_to_control_signle_turn_position := can.Data{0xA6, 0x00, 0x00, 0x02, s_delt_c_deg[0], s_delt_c_deg[1], 0x00, 0x00} // does not stop

	s_data_to_read_multi_turn_motor_position := []can.Data{
		data_to_read_multi_turn_encoder_position_data,
		data_to_read_multi_turn_encoder_position_data,
	}

	// s_data_to_set_multi_turn_position_zero_to_rom_RESTART_NEEDED := []can.Data{
	// 	data_to_set_multi_turn_position_zero_to_rom_RESTART_NEEDED,
	// 	data_to_set_multi_turn_position_zero_to_rom_RESTART_NEEDED,
	// }

	s_data_to_torque_closed_loop_control := []can.Data{
		data_to_torque_closed_loop_control_cw,
		data_to_torque_closed_loop_control_ccw,
	}

	s_data_to_shutdown_motor := []can.Data{
		data_to_shutdown_motor,
		data_to_shutdown_motor,
	}

	s_data_to_control_absolute_position_in_closed_loop_0 := []can.Data{
		data_to_control_absolute_position_in_closed_loop_0,
		data_to_control_absolute_position_in_closed_loop_0,
	}

	s_data_to_control_absolute_position_in_closed_loop_360 := []can.Data{
		data_to_control_absolute_position_in_closed_loop_360,
		data_to_control_absolute_position_in_closed_loop_360,
	}

	slice_of_sequance := [][]can.Data{
		s_data_to_read_multi_turn_motor_position,
		// s_data_to_set_multi_turn_position_zero_to_rom_RESTART_NEEDED,
		// s_data_to_shutdown_motor,
		s_data_to_torque_closed_loop_control,
		s_data_to_control_absolute_position_in_closed_loop_0,
		s_data_to_control_absolute_position_in_closed_loop_360,
		s_data_to_shutdown_motor,
	}

	// data := can.Data{0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	for i, s_data := range slice_of_sequance {
		err = send_command_to_each_motors(ctx, pTx, motor_ids, s_data)
		if err != nil {
			log.Println("error when sending command_to_each_motors...sequence:", i)
			log.Fatal(err)
		}
		log.Println("Waiting for a second ...")
		time.Sleep(time.Second)
	}

	log.Println("wating 1 sec before exiting main() to confirm othe processes will finish properly")
	time.Sleep(time.Second)
	log.Println("Exiting main()")

	// fmt.Printf("Get message: %v\n", msg)
}
