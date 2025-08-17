import re

# Read the file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\public_device_routes.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Add a new GetDevices method after GetDeviceInfo
get_device_info_pattern = r'(// GetDeviceInfo returns device information\nfunc \(api \*PublicDeviceAPI\) GetDeviceInfo\(c \*fiber\.Ctx\) error \{[\s\S]*?\n\}\n)'

new_get_devices_method = '''
// GetDevices returns all devices (for public view, returns only the current device)
func (api *PublicDeviceAPI) GetDevices(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Device not found"})
	}
	
	// For public view, only return the current device
	devices := []fiber.Map{
		{
			"id":           device.ID,
			"device_name":  device.DeviceName,
			"phone":        device.Phone,
			"jid":          device.Phone,
			"status":       device.Status,
			"created_at":   device.CreatedAt,
		},
	}
	
	return c.JSON(fiber.Map{
		"code": "SUCCESS",
		"results": devices,
	})
}
'''

# Insert the new method after GetDeviceInfo
content = re.sub(get_device_info_pattern, r'\1' + new_get_devices_method + '\n', content)

# Update the route mapping - change GetDeviceInfo to GetDevices
content = re.sub(
    r'publicAPI\.Get\("/devices", api\.GetDeviceInfo\)',
    'publicAPI.Get("/devices", api.GetDevices)',
    content
)

# Write back
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\public_device_routes.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("Added GetDevices method and updated route!")
