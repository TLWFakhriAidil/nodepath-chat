import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { 
  CreditCard, 
  Calendar,
  DollarSign,
  Download,
  ExternalLink,
  CheckCircle,
  Clock,
  AlertCircle
} from 'lucide-react';

/**
 * Billing page component
 * Displays current subscription, payment methods, billing history and Billplz integration
 */
const Billing = () => {
  // Mock data for demonstration
  const currentPlan = {
    name: 'Pro Plan',
    price: 'RM 99.00',
    period: 'monthly',
    status: 'active',
    nextBilling: '2025-11-06',
    features: [
      'Unlimited flows',
      'Advanced analytics',
      'Priority support',
      'Custom integrations'
    ]
  };

  const paymentHistory = [
    {
      id: 'inv_001',
      date: '2025-10-06',
      amount: 'RM 99.00',
      status: 'paid',
      description: 'Pro Plan - Monthly subscription'
    },
    {
      id: 'inv_002', 
      date: '2025-09-06',
      amount: 'RM 99.00',
      status: 'paid',
      description: 'Pro Plan - Monthly subscription'
    },
    {
      id: 'inv_003',
      date: '2025-08-06',
      amount: 'RM 99.00',
      status: 'paid',
      description: 'Pro Plan - Monthly subscription'
    }
  ];

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

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col space-y-2">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Billing & Subscription
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Manage your subscription, payment methods, and billing history
        </p>
      </div>

      {/* Current Subscription */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
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
                  <h3 className="text-lg font-semibold">{currentPlan.name}</h3>
                  <p className="text-gray-600 dark:text-gray-400">
                    {currentPlan.price} / {currentPlan.period}
                  </p>
                </div>
                {getStatusBadge(currentPlan.status)}
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                    Next billing date
                  </p>
                  <div className="flex items-center space-x-2">
                    <Calendar className="w-4 h-4 text-gray-400" />
                    <span className="text-sm">{currentPlan.nextBilling}</span>
                  </div>
                </div>
                
                <div className="space-y-2">
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                    Amount due
                  </p>
                  <div className="flex items-center space-x-2">
                    <DollarSign className="w-4 h-4 text-gray-400" />
                    <span className="text-sm">{currentPlan.price}</span>
                  </div>
                </div>
              </div>

              <div className="pt-4">
                <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
                  Plan features
                </p>
                <ul className="space-y-1">
                  {currentPlan.features.map((feature, index) => (
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
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Payment Integration */}
        <div>
          <Card>
            <CardHeader>
              <CardTitle>Payment Integration</CardTitle>
              <CardDescription>
                Secure payments powered by Billplz
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-center p-6 border-2 border-dashed border-gray-200 dark:border-gray-700 rounded-lg">
                <div className="text-center space-y-2">
                  <CreditCard className="w-8 h-8 mx-auto text-gray-400" />
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    Billplz Integration Ready
                  </p>
                </div>
              </div>
              
              <Button className="w-full" variant="outline">
                <ExternalLink className="w-4 h-4 mr-2" />
                View Billplz Integration
              </Button>
              
              <div className="text-xs text-gray-500 dark:text-gray-400">
                <p>✓ Secure payment processing</p>
                <p>✓ Malaysian payment methods</p>
                <p>✓ Automated billing</p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

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
            {paymentHistory.map((payment) => (
              <div key={payment.id}>
                <div className="flex items-center justify-between py-3">
                  <div className="flex items-center space-x-4">
                    {getStatusIcon(payment.status)}
                    <div>
                      <p className="text-sm font-medium">{payment.description}</p>
                      <p className="text-xs text-gray-500 dark:text-gray-400">
                        Invoice #{payment.id} • {payment.date}
                      </p>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-3">
                    <span className="text-sm font-medium">{payment.amount}</span>
                    {getStatusBadge(payment.status)}
                    <Button variant="ghost" size="sm">
                      <Download className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
                {payment.id !== paymentHistory[paymentHistory.length - 1].id && (
                  <Separator />
                )}
              </div>
            ))}
          </div>
          
          <div className="pt-4">
            <Button variant="outline" className="w-full">
              View All Invoices
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Payment Methods */}
      <Card>
        <CardHeader>
          <CardTitle>Payment Methods</CardTitle>
          <CardDescription>
            Manage your payment methods for automatic billing
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center p-8 border-2 border-dashed border-gray-200 dark:border-gray-700 rounded-lg">
            <div className="text-center space-y-3">
              <CreditCard className="w-12 h-12 mx-auto text-gray-400" />
              <div>
                <p className="text-sm font-medium text-gray-900 dark:text-white">
                  No payment methods added
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  Add a payment method for automatic billing
                </p>
              </div>
              <Button>
                Add Payment Method
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default Billing;