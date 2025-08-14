package fixes

// This file contains the fix for the DeleteDevice function
// The issue is in the GetDevice function where it's not handling NULL values properly

// The fix should be applied to src/repository/user_repository.go
// In the GetDevice function, change:
//
// err := r.db.QueryRow(query, userID, deviceID).Scan(
//     &device.ID, &device.UserID, &device.DeviceName,
//     &phone, &jid, &device.Status,
//     &device.CreatedAt, &updatedAt, &device.LastSeen,
// )
//
// The issue is that the query might return devices where the user_id doesn't match
// We need to fix the query to only check device ID since DeleteDevice doesn't pass userID

// Fixed GetDevice function for deletion:
func GetDeviceByID(deviceID string) (*models.UserDevice, error) {
    query := `
        SELECT id, user_id, device_name, COALESCE(phone, ''), COALESCE(jid, ''), status, created_at, updated_at, last_seen
        FROM user_devices
        WHERE id = $1
    `
    
    var device models.UserDevice
    var updatedAt sql.NullTime
    
    err := r.db.QueryRow(query, deviceID).Scan(
        &device.ID, &device.UserID, &device.DeviceName,
        &device.Phone, &device.JID, &device.Status,
        &device.CreatedAt, &updatedAt, &device.LastSeen,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("device not found")
    }
    
    if err != nil {
        return nil, fmt.Errorf("failed to get device: %w", err)
    }
    
    if updatedAt.Valid {
        device.UpdatedAt = updatedAt.Time
    }
    
    return &device, nil
}
