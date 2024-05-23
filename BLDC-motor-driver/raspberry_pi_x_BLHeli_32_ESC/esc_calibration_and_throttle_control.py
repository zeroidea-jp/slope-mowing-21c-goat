import time
import os
from gpiozero import PWMOutputDevice, Button, OutputDevice

# ESC settings
ESC_PIN = 12  # GPIO pin connected to the ESC signal wire
ESC_MIN_DUTY = 0.05  # Minimum duty cycle for the ESC
ESC_MAX_DUTY = 0.10  # Maximum duty cycle for the ESC

# Relay settings (intended for raspberry pi 2/3)
RELAY_PIN = 26  # GPIO pin connected to the relay

# Initialize PWM output
esc = PWMOutputDevice(ESC_PIN, frequency=50, initial_value=0)

# Initialize relay output
relay = OutputDevice(RELAY_PIN, initial_value=True)

# Initialize buttons
up_button = Button(21) 
down_button = Button(20)
quit_button = Button(16)

def set_esc_duty(duty_cycle):
    esc.value = duty_cycle

def calibrate_esc():
    print("Calibrating ESC...")
    set_esc_duty(ESC_MAX_DUTY)
    time.sleep(2)
    set_esc_duty(ESC_MIN_DUTY)
    time.sleep(2)
    print("ESC calibration complete.")

def main():
    # Turn on the relay to supply power to the BLDC motor
    relay.on()
    print("Relay turned on. Power supplied to the BLDC motor.")

    # Calibrate the ESC
    calibrate_esc()

    # Set initial throttle to minimum
    current_duty = ESC_MIN_DUTY
    set_esc_duty(current_duty)

    # Main loop
    while True:
        if up_button.is_pressed:
            if current_duty < ESC_MAX_DUTY:
                current_duty += 0.01
                set_esc_duty(current_duty)
                print(f"Throttle increased. Current duty: {current_duty:.2f}")
            else:
                print("Throttle is already at maximum.")
            time.sleep(0.2)  # Debounce delay
        
        elif down_button.is_pressed:
            if current_duty > ESC_MIN_DUTY:
                current_duty -= 0.01
                set_esc_duty(current_duty)
                print(f"Throttle decreased. Current duty: {current_duty:.2f}")
            else:
                print("Throttle is already at minimum.")
            time.sleep(0.2)  # Debounce delay
        
        elif quit_button.is_pressed:
            break

    # Clean up
    set_esc_duty(ESC_MIN_DUTY)
    esc.close()
    relay.off()
    print("Motor stopped, PWM output closed, and relay turned off.")
    
    # Shutdown the Raspberry Pi
    print("Shutting down the Raspberry Pi...")
    os.system("sudo shutdown -h now")

if __name__ == "__main__":
    main()