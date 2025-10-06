import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { 
  CreditCard, 
  Calendar,
  DollarSign,
  Download,
  CheckCircle,
  Clock,
  AlertCircle,
  Loader2
} from 'lucide-react';

// Types for billing data
interface Subscription {
  id: string;
  plan_name: string;
  plan_price: number;
  plan_period: string;
  status: string;
  next_billing_date: string;
  features: string[];
}

interface BillingHistoryItem {
  id: string;
  invoice_number: string;
  amount: number;
  currency: string;
  description: string;
  status: string;
  payment_date: string | null;
  created_at: string;
}

interface BillingData {
  subscription: Subscription;
  billing_history: BillingHistoryItem[];
  total_count: number;
}

/**
 * Billing page component
 * Displays current subscription and billing history with real API integration
 */
const Billing = () => {
  const [billingData, setBillingData] = useState<BillingData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [testingPayment, setTestingPayment] = useState(false);

  // Fetch billing data from API
  useEffect(() => {
    fetchBillingData();
  }, []);

  const fetchBillingData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch('/api/billing/', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          // Add authentication header if needed
          // 'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        // Handle 500 errors gracefully (database issues)
        if (response.status === 500) {
          console.warn('Billing API returned 500 error, using fallback data');
          const fallbackData = {
            success: true,
            data: {
              subscription: {
                id: 'fallback_sub_001',
                plan_name: 'Default Plan (Database Error)',
                plan_price: 1.00,
                plan_period: 'monthly',
                status: 'active',
                next_billing_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
                features: ['WhatsApp Bot Integration', 'Flow Builder Access', 'Basic Analytics'],
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString()
              },
              billing_history: [
                {
                  id: 'fallback_hist_001',
                  invoice_number: 'INV-FALLBACK-001',
                  amount: 1.00,
                  currency: 'MYR',
                  description: 'Default Plan - Please apply database migrations',
                  status: 'paid',
                  payment_date: new Date().toISOString().split('T')[0],
                  created_at: new Date().toISOString()
                }
              ],
              pagination: {
                total: 1,
                limit: 10,
                offset: 0,
                has_more: false
              }
            }
          };
          setBillingData(fallbackData);
          setError('Billing data partially unavailable. Please apply database migrations for full functionality.');
          return;
        }
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      setBillingData(data);
    } catch (err) {
      console.error('Failed to fetch billing data:', err);
      setError('Failed to load billing data. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Test payment function
  const handleTestPayment = async () => {
    try {
      setTestingPayment(true);
      
      const response = await fetch('/api/billing/test-payment', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          // Add authentication header if needed
          // 'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      
      // Show success message
      alert(`Test payment created successfully! Payment ID: ${result.payment?.id}`);
      
      // Refresh billing data
      fetchBillingData();
    } catch (err) {
      console.error('Failed to create test payment:', err);
      alert('Failed to create test payment. Please try again.');
    } finally {
      setTestingPayment(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'paid':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'pending':
        return <Clock className="w-4 h-4 text-yellow-500" />;
      case 'failed':
        return <AlertCircle className="w-4 h-4 text-red-500" />;
      default:
        return <Clock className="w-4 h-4 text-gray-500" />;
    }
  };

  const getStatusBadge = (status: string) => {
    const variants = {
      paid: 'default',
      pending: 'secondary',
      failed: 'destructive',
      active: 'default'
    } as const;

    return (
      <Badge variant={variants[status as keyof typeof variants] || 'secondary'}>
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </Badge>
    );
  };

  // Loading state
  if (loading) {
    return (
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        <div className="flex flex-col space-y-2">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Billing & Subscription
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage your subscription and view billing history
          </p>
        </div>
        
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="flex items-center space-x-2">
            <Loader2 className="w-6 h-6 animate-spin" />
            <span>Loading billing data...</span>
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        <div className="flex flex-col space-y-2">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Billing & Subscription
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage your subscription and view billing history
          </p>
        </div>
        
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center space-x-2 text-red-600">
              <AlertCircle className="w-5 h-5" />
              <span>{error}</span>
            </div>
            <Button 
              onClick={fetchBillingData} 
              className="mt-4"
              variant="outline"
            >
              Try Again
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // No data state
  if (!billingData) {
    return (
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        <div className="flex flex-col space-y-2">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Billing & Subscription
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage your subscription and view billing history
          </p>
        </div>
        
        <Card>
          <CardContent className="p-6">
            <p>No billing data available.</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const { subscription, billing_history } = billingData;

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col space-y-2">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Billing & Subscription
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Manage your subscription and view billing history
        </p>
      </div>

      {/* Current Subscription */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <CreditCard className="w-5 h-5" />
            <span>Current Subscription</span>
          </CardTitle>
          <CardDescription>
            Your active subscription plan and billing details
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold">{subscription.plan_name}</h3>
              <p className="text-gray-600 dark:text-gray-400">
                RM {subscription.plan_price.toFixed(2)} / {subscription.plan_period}
              </p>
            </div>
            {getStatusBadge(subscription.status)}
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Next billing date
              </p>
              <div className="flex items-center space-x-2">
                <Calendar className="w-4 h-4 text-gray-400" />
                <span className="text-sm">
                  {new Date(subscription.next_billing_date).toLocaleDateString()}
                </span>
              </div>
            </div>
            
            <div className="space-y-2">
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Amount due
              </p>
              <div className="flex items-center space-x-2">
                <DollarSign className="w-4 h-4 text-gray-400" />
                <span className="text-sm">RM {subscription.plan_price.toFixed(2)}</span>
              </div>
            </div>
          </div>

          <div className="pt-4">
            <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
              Plan features
            </p>
            <ul className="space-y-1">
              {subscription.features.map((feature, index) => (
                <li key={index} className="flex items-center space-x-2 text-sm">
                  <CheckCircle className="w-3 h-3 text-green-500" />
                  <span>{feature}</span>
                </li>
              ))}
            </ul>
          </div>

          <div className="flex space-x-3 pt-4">
            <Button variant="outline">
              Change Plan
            </Button>
            <Button variant="outline">
              Cancel Subscription
            </Button>
            <Button 
              onClick={handleTestPayment}
              disabled={testingPayment}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {testingPayment ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Testing...
                </>
              ) : (
                'Test RM 1.00 Payment'
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Billing History */}
      <Card>
        <CardHeader>
          <CardTitle>Billing History</CardTitle>
          <CardDescription>
            View and download your payment history and invoices
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {billing_history.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-gray-500 dark:text-gray-400">No billing history available.</p>
              </div>
            ) : (
              billing_history.map((payment, index) => (
                <div key={payment.id}>
                  <div className="flex items-center justify-between py-3">
                    <div className="flex items-center space-x-4">
                      {getStatusIcon(payment.status)}
                      <div>
                        <p className="text-sm font-medium">{payment.description}</p>
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          Invoice #{payment.invoice_number} • {new Date(payment.created_at).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                    
                    <div className="flex items-center space-x-3">
                      <span className="text-sm font-medium">
                        {payment.currency} {payment.amount.toFixed(2)}
                      </span>
                      {getStatusBadge(payment.status)}
                      <Button variant="ghost" size="sm">
                        <Download className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                  {index !== billing_history.length - 1 && (
                    <Separator />
                  )}
                </div>
              ))
            )}
          </div>
          
          <div className="pt-4">
            <Button variant="outline" className="w-full">
              View All Invoices
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default Billing;