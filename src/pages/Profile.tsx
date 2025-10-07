import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { User, Mail, Phone, Calendar, Shield, Save, Eye, EyeOff } from 'lucide-react';

interface User {
  id: string;
  email: string;
  full_name: string;
  gmail?: string;
  phone?: string;
  status: string;
  expired?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login?: string;
}

interface ProfileFormData {
  full_name: string;
  gmail?: string;
  phone?: string;
  password?: string;
  new_password?: string;
}

const Profile: React.FC = () => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [showPasswordFields, setShowPasswordFields] = useState(false);
  const [formData, setFormData] = useState<ProfileFormData>({
    full_name: '',
    gmail: '',
    phone: '',
    password: '',
    new_password: '',
  });
  const [alert, setAlert] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  useEffect(() => {
    fetchProfile();
  }, []);

  const fetchProfile = async () => {
    try {
      const response = await fetch('/api/profile/', {
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to fetch profile');
      }

      const result = await response.json();
      if (result.success) {
        setUser(result.data);
        setFormData({
          full_name: result.data.full_name,
          gmail: result.data.gmail || '',
          phone: result.data.phone || '',
          password: '',
          new_password: '',
        });
      } else {
        throw new Error(result.error || 'Failed to fetch profile');
      }
    } catch (error) {
      console.error('Error fetching profile:', error);
      setAlert({ type: 'error', message: 'Failed to load profile data' });
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSave = async () => {
    setSaving(true);
    setAlert(null);

    try {
      // Prepare payload - only send fields that have values
      const payload: any = {
        full_name: formData.full_name,
      };

      if (formData.gmail) {
        payload.gmail = formData.gmail;
      }

      if (formData.phone) {
        payload.phone = formData.phone;
      }

      // Handle password change
      if (formData.password && formData.new_password) {
        payload.password = formData.password;
        payload.new_password = formData.new_password;
      }

      const response = await fetch('/api/profile/', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(payload),
      });

      const result = await response.json();

      if (result.success) {
        setUser(result.data);
        setIsEditing(false);
        setShowPasswordFields(false);
        setFormData(prev => ({
          ...prev,
          password: '',
          new_password: '',
        }));
        setAlert({ type: 'success', message: 'Profile updated successfully!' });
      } else {
        throw new Error(result.error || 'Failed to update profile');
      }
    } catch (error) {
      console.error('Error updating profile:', error);
      setAlert({ type: 'error', message: error instanceof Error ? error.message : 'Failed to update profile' });
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    if (user) {
      setFormData({
        full_name: user.full_name,
        gmail: user.gmail || '',
        phone: user.phone || '',
        password: '',
        new_password: '',
      });
    }
    setIsEditing(false);
    setShowPasswordFields(false);
    setAlert(null);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'trial':
        return 'bg-blue-100 text-blue-800';
      case 'expired':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return (
      <div className=\"flex items-center justify-center h-96\">
        <div className=\"animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500\"></div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className=\"flex items-center justify-center h-96\">
        <Alert>
          <AlertDescription>Failed to load profile data. Please try refreshing the page.</AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className=\"container mx-auto p-6 max-w-4xl\">
      <div className=\"space-y-6\">
        {/* Header */}
        <div className=\"flex items-center justify-between\">
          <div>
            <h1 className=\"text-2xl font-bold\">Profile</h1>
            <p className=\"text-gray-600\">Manage your account settings and preferences</p>
          </div>
          {!isEditing && (
            <Button onClick={() => setIsEditing(true)} className=\"flex items-center gap-2\">
              <User className=\"h-4 w-4\" />
              Edit Profile
            </Button>
          )}
        </div>\n\n        {/* Alert */}\n        {alert && (\n          <Alert className={alert.type === 'error' ? 'border-red-500 bg-red-50' : 'border-green-500 bg-green-50'}>\n            <AlertDescription className={alert.type === 'error' ? 'text-red-700' : 'text-green-700'}>\n              {alert.message}\n            </AlertDescription>\n          </Alert>\n        )}\n\n        <div className=\"grid gap-6 md:grid-cols-2\">\n          {/* Profile Information */}\n          <Card>\n            <CardHeader>\n              <CardTitle className=\"flex items-center gap-2\">\n                <User className=\"h-5 w-5\" />\n                Profile Information\n              </CardTitle>\n              <CardDescription>\n                Your personal information and contact details\n              </CardDescription>\n            </CardHeader>\n            <CardContent className=\"space-y-4\">\n              <div className=\"space-y-2\">\n                <Label htmlFor=\"full_name\">Full Name</Label>\n                <Input\n                  id=\"full_name\"\n                  name=\"full_name\"\n                  value={formData.full_name}\n                  onChange={handleInputChange}\n                  disabled={!isEditing}\n                  placeholder=\"Enter your full name\"\n                />\n              </div>\n\n              <div className=\"space-y-2\">\n                <Label htmlFor=\"email\">Email</Label>\n                <Input\n                  id=\"email\"\n                  value={user.email}\n                  disabled\n                  className=\"bg-gray-50\"\n                />\n                <p className=\"text-sm text-gray-500\">Email cannot be changed</p>\n              </div>\n\n              <div className=\"space-y-2\">\n                <Label htmlFor=\"gmail\">Gmail (Optional)</Label>\n                <Input\n                  id=\"gmail\"\n                  name=\"gmail\"\n                  type=\"email\"\n                  value={formData.gmail}\n                  onChange={handleInputChange}\n                  disabled={!isEditing}\n                  placeholder=\"Enter your Gmail address\"\n                />\n              </div>\n\n              <div className=\"space-y-2\">\n                <Label htmlFor=\"phone\">Phone Number (Optional)</Label>\n                <Input\n                  id=\"phone\"\n                  name=\"phone\"\n                  value={formData.phone}\n                  onChange={handleInputChange}\n                  disabled={!isEditing}\n                  placeholder=\"Enter your phone number\"\n                />\n              </div>\n            </CardContent>\n          </Card>\n\n          {/* Account Status */}\n          <Card>\n            <CardHeader>\n              <CardTitle className=\"flex items-center gap-2\">\n                <Shield className=\"h-5 w-5\" />\n                Account Status\n              </CardTitle>\n              <CardDescription>\n                Your account information and status\n              </CardDescription>\n            </CardHeader>\n            <CardContent className=\"space-y-4\">\n              <div className=\"flex items-center justify-between\">\n                <span className=\"text-sm font-medium\">Status:</span>\n                <Badge className={getStatusColor(user.status)}>\n                  {user.status}\n                </Badge>\n              </div>\n\n              {user.expired && (\n                <div className=\"flex items-center justify-between\">\n                  <span className=\"text-sm font-medium\">Expires:</span>\n                  <span className=\"text-sm text-gray-600\">{formatDate(user.expired)}</span>\n                </div>\n              )}\n\n              <div className=\"flex items-center justify-between\">\n                <span className=\"text-sm font-medium\">Member Since:</span>\n                <span className=\"text-sm text-gray-600\">{formatDate(user.created_at)}</span>\n              </div>\n\n              {user.last_login && (\n                <div className=\"flex items-center justify-between\">\n                  <span className=\"text-sm font-medium\">Last Login:</span>\n                  <span className=\"text-sm text-gray-600\">{formatDate(user.last_login)}</span>\n                </div>\n              )}\n\n              <div className=\"flex items-center justify-between\">\n                <span className=\"text-sm font-medium\">Account ID:</span>\n                <span className=\"text-sm text-gray-600 font-mono\">{user.id.substring(0, 8)}...</span>\n              </div>\n            </CardContent>\n          </Card>\n        </div>\n\n        {/* Password Change Section */}\n        {isEditing && (\n          <Card>\n            <CardHeader>\n              <CardTitle className=\"flex items-center gap-2\">\n                <Shield className=\"h-5 w-5\" />\n                Change Password\n              </CardTitle>\n              <CardDescription>\n                Update your account password (optional)\n              </CardDescription>\n            </CardHeader>\n            <CardContent>\n              <div className=\"space-y-4\">\n                <Button\n                  type=\"button\"\n                  variant=\"outline\"\n                  onClick={() => setShowPasswordFields(!showPasswordFields)}\n                  className=\"flex items-center gap-2\"\n                >\n                  {showPasswordFields ? (\n                    <>\n                      <EyeOff className=\"h-4 w-4\" />\n                      Hide Password Fields\n                    </>\n                  ) : (\n                    <>\n                      <Eye className=\"h-4 w-4\" />\n                      Change Password\n                    </>\n                  )}\n                </Button>\n\n                {showPasswordFields && (\n                  <div className=\"grid gap-4 md:grid-cols-2\">\n                    <div className=\"space-y-2\">\n                      <Label htmlFor=\"password\">Current Password</Label>\n                      <Input\n                        id=\"password\"\n                        name=\"password\"\n                        type=\"password\"\n                        value={formData.password}\n                        onChange={handleInputChange}\n                        placeholder=\"Enter current password\"\n                      />\n                    </div>\n                    <div className=\"space-y-2\">\n                      <Label htmlFor=\"new_password\">New Password</Label>\n                      <Input\n                        id=\"new_password\"\n                        name=\"new_password\"\n                        type=\"password\"\n                        value={formData.new_password}\n                        onChange={handleInputChange}\n                        placeholder=\"Enter new password\"\n                      />\n                    </div>\n                  </div>\n                )}\n              </div>\n            </CardContent>\n          </Card>\n        )}\n\n        {/* Action Buttons */}\n        {isEditing && (\n          <div className=\"flex justify-end gap-3\">\n            <Button variant=\"outline\" onClick={handleCancel}>\n              Cancel\n            </Button>\n            <Button onClick={handleSave} disabled={saving} className=\"flex items-center gap-2\">\n              <Save className=\"h-4 w-4\" />\n              {saving ? 'Saving...' : 'Save Changes'}\n            </Button>\n          </div>\n        )}\n      </div>\n    </div>\n  );\n};\n\nexport default Profile;"