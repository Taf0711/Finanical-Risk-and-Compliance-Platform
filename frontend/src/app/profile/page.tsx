'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { 
  UserCircleIcon, 
  CogIcon, 
  ShieldCheckIcon,
  BellIcon,
  EyeIcon,
  EyeSlashIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon
} from '@heroicons/react/24/outline';
import { useRequireAuth } from '@/hooks/useAuth';
import DashboardLayout from '@/components/DashboardLayout';

const profileSchema = z.object({
  first_name: z.string().min(1, 'First name is required'),
  last_name: z.string().min(1, 'Last name is required'),
  email: z.string().email('Invalid email address'),
});

const passwordSchema = z.object({
  current_password: z.string().min(6, 'Current password is required'),
  new_password: z.string().min(8, 'New password must be at least 8 characters'),
  confirm_password: z.string(),
}).refine((data) => data.new_password === data.confirm_password, {
  message: "Passwords don't match",
  path: ["confirm_password"],
});

const preferencesSchema = z.object({
  notifications_email: z.boolean(),
  notifications_push: z.boolean(),
  notifications_sms: z.boolean(),
  risk_alerts: z.boolean(),
  price_alerts: z.boolean(),
  portfolio_updates: z.boolean(),
  theme: z.enum(['light', 'dark', 'system']),
  timezone: z.string(),
  currency: z.string(),
});

type ProfileFormData = z.infer<typeof profileSchema>;
type PasswordFormData = z.infer<typeof passwordSchema>;
type PreferencesFormData = z.infer<typeof preferencesSchema>;

export default function ProfilePage() {
  const { user, loading } = useRequireAuth();
  const [activeTab, setActiveTab] = useState('profile');
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateSuccess, setUpdateSuccess] = useState('');
  const [updateError, setUpdateError] = useState('');

  const profileForm = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      first_name: user?.first_name || '',
      last_name: user?.last_name || '',
      email: user?.email || '',
    }
  });

  const passwordForm = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
  });

  const preferencesForm = useForm<PreferencesFormData>({
    resolver: zodResolver(preferencesSchema),
    defaultValues: {
      notifications_email: true,
      notifications_push: true,
      notifications_sms: false,
      risk_alerts: true,
      price_alerts: true,
      portfolio_updates: true,
      theme: 'dark',
      timezone: 'UTC',
      currency: 'USD',
    }
  });

  const onProfileSubmit = async (data: ProfileFormData) => {
    setIsUpdating(true);
    setUpdateError('');
    setUpdateSuccess('');
    
    try {
      // TODO: Implement API call to update profile
      await new Promise(resolve => setTimeout(resolve, 1000)); // Mock API call
      setUpdateSuccess('Profile updated successfully!');
    } catch (error) {
      setUpdateError('Failed to update profile. Please try again.');
    } finally {
      setIsUpdating(false);
    }
  };

  const onPasswordSubmit = async (data: PasswordFormData) => {
    setIsUpdating(true);
    setUpdateError('');
    setUpdateSuccess('');
    
    try {
      // TODO: Implement API call to change password
      await new Promise(resolve => setTimeout(resolve, 1000)); // Mock API call
      setUpdateSuccess('Password changed successfully!');
      passwordForm.reset();
    } catch (error) {
      setUpdateError('Failed to change password. Please try again.');
    } finally {
      setIsUpdating(false);
    }
  };

  const onPreferencesSubmit = async (data: PreferencesFormData) => {
    setIsUpdating(true);
    setUpdateError('');
    setUpdateSuccess('');
    
    try {
      // TODO: Implement API call to update preferences
      await new Promise(resolve => setTimeout(resolve, 1000)); // Mock API call
      setUpdateSuccess('Preferences updated successfully!');
    } catch (error) {
      setUpdateError('Failed to update preferences. Please try again.');
    } finally {
      setIsUpdating(false);
    }
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-96">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      </DashboardLayout>
    );
  }

  const tabs = [
    { id: 'profile', label: 'Profile', icon: UserCircleIcon },
    { id: 'security', label: 'Security', icon: ShieldCheckIcon },
    { id: 'preferences', label: 'Preferences', icon: CogIcon },
    { id: 'notifications', label: 'Notifications', icon: BellIcon },
  ];

  return (
    <DashboardLayout>
      <div className="p-6 max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">Account Settings</h1>
          <p className="text-slate-400 mt-2">Manage your account settings and preferences</p>
        </div>

        {/* Success/Error Messages */}
        {updateSuccess && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 p-4 bg-green-500/20 border border-green-500/30 rounded-lg text-green-200 flex items-center"
          >
            <CheckCircleIcon className="h-5 w-5 mr-2" />
            {updateSuccess}
          </motion.div>
        )}

        {updateError && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 p-4 bg-red-500/20 border border-red-500/30 rounded-lg text-red-200 flex items-center"
          >
            <ExclamationTriangleIcon className="h-5 w-5 mr-2" />
            {updateError}
          </motion.div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Sidebar Navigation */}
          <div className="lg:col-span-1">
            <nav className="space-y-2">
              {tabs.map((tab) => {
                const Icon = tab.icon;
                return (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id)}
                    className={`w-full flex items-center px-4 py-3 rounded-lg text-left transition-colors ${
                      activeTab === tab.id
                        ? 'bg-blue-600 text-white'
                        : 'text-slate-400 hover:text-white hover:bg-slate-700'
                    }`}
                  >
                    <Icon className="h-5 w-5 mr-3" />
                    {tab.label}
                  </button>
                );
              })}
            </nav>
          </div>

          {/* Main Content */}
          <div className="lg:col-span-3">
            <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6">
              {activeTab === 'profile' && (
                <div>
                  <h2 className="text-xl font-semibold text-white mb-6">Profile Information</h2>
                  <form onSubmit={profileForm.handleSubmit(onProfileSubmit)} className="space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          First Name
                        </label>
                        <input
                          {...profileForm.register('first_name')}
                          type="text"
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        {profileForm.formState.errors.first_name && (
                          <p className="mt-1 text-sm text-red-400">
                            {profileForm.formState.errors.first_name.message}
                          </p>
                        )}
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          Last Name
                        </label>
                        <input
                          {...profileForm.register('last_name')}
                          type="text"
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        {profileForm.formState.errors.last_name && (
                          <p className="mt-1 text-sm text-red-400">
                            {profileForm.formState.errors.last_name.message}
                          </p>
                        )}
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-slate-300 mb-2">
                        Email Address
                      </label>
                      <input
                        {...profileForm.register('email')}
                        type="email"
                        className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                      {profileForm.formState.errors.email && (
                        <p className="mt-1 text-sm text-red-400">
                          {profileForm.formState.errors.email.message}
                        </p>
                      )}
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={isUpdating}
                        className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                      >
                        {isUpdating ? 'Updating...' : 'Update Profile'}
                      </button>
                    </div>
                  </form>
                </div>
              )}

              {activeTab === 'security' && (
                <div>
                  <h2 className="text-xl font-semibold text-white mb-6">Change Password</h2>
                  <form onSubmit={passwordForm.handleSubmit(onPasswordSubmit)} className="space-y-6">
                    <div>
                      <label className="block text-sm font-medium text-slate-300 mb-2">
                        Current Password
                      </label>
                      <div className="relative">
                        <input
                          {...passwordForm.register('current_password')}
                          type={showCurrentPassword ? 'text' : 'password'}
                          className="w-full px-4 py-3 pr-12 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        <button
                          type="button"
                          onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                          className="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-300"
                        >
                          {showCurrentPassword ? <EyeSlashIcon className="h-5 w-5" /> : <EyeIcon className="h-5 w-5" />}
                        </button>
                      </div>
                      {passwordForm.formState.errors.current_password && (
                        <p className="mt-1 text-sm text-red-400">
                          {passwordForm.formState.errors.current_password.message}
                        </p>
                      )}
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-slate-300 mb-2">
                        New Password
                      </label>
                      <div className="relative">
                        <input
                          {...passwordForm.register('new_password')}
                          type={showNewPassword ? 'text' : 'password'}
                          className="w-full px-4 py-3 pr-12 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        <button
                          type="button"
                          onClick={() => setShowNewPassword(!showNewPassword)}
                          className="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-300"
                        >
                          {showNewPassword ? <EyeSlashIcon className="h-5 w-5" /> : <EyeIcon className="h-5 w-5" />}
                        </button>
                      </div>
                      {passwordForm.formState.errors.new_password && (
                        <p className="mt-1 text-sm text-red-400">
                          {passwordForm.formState.errors.new_password.message}
                        </p>
                      )}
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-slate-300 mb-2">
                        Confirm New Password
                      </label>
                      <div className="relative">
                        <input
                          {...passwordForm.register('confirm_password')}
                          type={showConfirmPassword ? 'text' : 'password'}
                          className="w-full px-4 py-3 pr-12 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        <button
                          type="button"
                          onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                          className="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-300"
                        >
                          {showConfirmPassword ? <EyeSlashIcon className="h-5 w-5" /> : <EyeIcon className="h-5 w-5" />}
                        </button>
                      </div>
                      {passwordForm.formState.errors.confirm_password && (
                        <p className="mt-1 text-sm text-red-400">
                          {passwordForm.formState.errors.confirm_password.message}
                        </p>
                      )}
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={isUpdating}
                        className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                      >
                        {isUpdating ? 'Changing...' : 'Change Password'}
                      </button>
                    </div>
                  </form>
                </div>
              )}

              {activeTab === 'preferences' && (
                <div>
                  <h2 className="text-xl font-semibold text-white mb-6">Preferences</h2>
                  <form onSubmit={preferencesForm.handleSubmit(onPreferencesSubmit)} className="space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          Theme
                        </label>
                        <select
                          {...preferencesForm.register('theme')}
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                          <option value="light">Light</option>
                          <option value="dark">Dark</option>
                          <option value="system">System</option>
                        </select>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          Timezone
                        </label>
                        <select
                          {...preferencesForm.register('timezone')}
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                          <option value="UTC">UTC</option>
                          <option value="America/New_York">Eastern Time</option>
                          <option value="America/Chicago">Central Time</option>
                          <option value="America/Denver">Mountain Time</option>
                          <option value="America/Los_Angeles">Pacific Time</option>
                        </select>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-slate-300 mb-2">
                        Default Currency
                      </label>
                      <select
                        {...preferencesForm.register('currency')}
                        className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                      >
                        <option value="USD">USD - US Dollar</option>
                        <option value="EUR">EUR - Euro</option>
                        <option value="GBP">GBP - British Pound</option>
                        <option value="JPY">JPY - Japanese Yen</option>
                        <option value="CAD">CAD - Canadian Dollar</option>
                      </select>
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={isUpdating}
                        className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                      >
                        {isUpdating ? 'Updating...' : 'Update Preferences'}
                      </button>
                    </div>
                  </form>
                </div>
              )}

              {activeTab === 'notifications' && (
                <div>
                  <h2 className="text-xl font-semibold text-white mb-6">Notification Settings</h2>
                  <form onSubmit={preferencesForm.handleSubmit(onPreferencesSubmit)} className="space-y-6">
                    <div className="space-y-4">
                      <h3 className="text-lg font-medium text-white">Notification Methods</h3>
                      
                      <div className="space-y-3">
                        <label className="flex items-center">
                          <input
                            {...preferencesForm.register('notifications_email')}
                            type="checkbox"
                            className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                          />
                          <span className="ml-3 text-slate-300">Email notifications</span>
                        </label>

                        <label className="flex items-center">
                          <input
                            {...preferencesForm.register('notifications_push')}
                            type="checkbox"
                            className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                          />
                          <span className="ml-3 text-slate-300">Push notifications</span>
                        </label>

                        <label className="flex items-center">
                          <input
                            {...preferencesForm.register('notifications_sms')}
                            type="checkbox"
                            className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                          />
                          <span className="ml-3 text-slate-300">SMS notifications</span>
                        </label>
                      </div>
                    </div>

                    <div className="space-y-4">
                      <h3 className="text-lg font-medium text-white">Alert Types</h3>
                      
                      <div className="space-y-3">
                        <label className="flex items-center">
                          <input
                            {...preferencesForm.register('risk_alerts')}
                            type="checkbox"
                            className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                          />
                          <span className="ml-3 text-slate-300">Risk alerts</span>
                        </label>

                        <label className="flex items-center">
                          <input
                            {...preferencesForm.register('price_alerts')}
                            type="checkbox"
                            className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                          />
                          <span className="ml-3 text-slate-300">Price alerts</span>
                        </label>

                        <label className="flex items-center">
                          <input
                            {...preferencesForm.register('portfolio_updates')}
                            type="checkbox"
                            className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                          />
                          <span className="ml-3 text-slate-300">Portfolio updates</span>
                        </label>
                      </div>
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={isUpdating}
                        className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                      >
                        {isUpdating ? 'Updating...' : 'Update Notifications'}
                      </button>
                    </div>
                  </form>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}
