import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';
import { CheckCircle, Clock, XCircle, Loader2, CreditCard } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import Swal from 'sweetalert2';

interface Order {
  id: number;
  amount: number;
  product: string;
  method: string;
  status: string;
  bill_id?: string;
  url?: string;
  created_at: string;
}

export default function Billings() {
  const [loading, setLoading] = useState(false);
  const [orders, setOrders] = useState<Order[]>([]);
  const [loadingOrders, setLoadingOrders] = useState(true);

  // Fetch user's orders
  useEffect(() => {
    fetchOrders();
  }, []);

  const fetchOrders = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        setLoadingOrders(false);
        return;
      }

      const response = await fetch('/api/billing/orders?limit=50&offset=0', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const data = await response.json();
        setOrders(data.orders || []);
      } else {
        console.error('Failed to fetch orders:', response.statusText);
      }
    } catch (error) {
      console.error('Failed to fetch orders:', error);
    } finally {
      setLoadingOrders(false);
    }
  };

  const handleUpgradeClick = async () => {
    setLoading(true);

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        toast.error('Please login first');
        setLoading(false);
        return;
      }

      const response = await fetch('/api/billing/create-order', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          amount: 1.00,
          product: 'Pro Plan - Monthly Subscription'
        }),
      });

      const data = await response.json();

      if (response.ok) {
        if (data.payment_url) {
          toast.success('Order created! Redirecting to payment...');
          // Redirect to Billplz payment page
          window.location.href = data.payment_url;
        } else {
          toast.success('Order created successfully!');
          fetchOrders(); // Refresh orders list
        }
      } else {
        // Check if profile is incomplete
        if (data.error === 'profile_incomplete') {
          await Swal.fire({
            icon: 'warning',
            title: 'Profile Incomplete',
            text: data.message || 'Please update your profile with both email and phone number before upgrading.',
            confirmButtonText: 'Go to Profile',
            confirmButtonColor: '#3085d6',
          });
          // Redirect to profile page
          window.location.href = '/profile';
        } else {
          toast.error(data.error || 'Failed to create order');
        }
      }
    } catch (error) {
      console.error('Error creating order:', error);
      toast.error('Failed to create order. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'Success':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'Processing':
        return <Clock className="w-4 h-4 text-blue-500" />;
      case 'Failed':
        return <XCircle className="w-4 h-4 text-red-500" />;
      default:
        return <Clock className="w-4 h-4 text-gray-500" />;
    }
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, any> = {
      'Success': 'default',
      'Processing': 'secondary',
      'Failed': 'destructive',
      'Pending': 'outline'
    };
    
    return (
      <Badge variant={variants[status] || 'outline'} className="flex items-center gap-1">
        {getStatusIcon(status)}
        {status}
      </Badge>
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
          Billing & Subscription
        </h2>
        <p className="text-muted-foreground">
          Manage your subscription and billing information
        </p>
      </div>

      {/* Current Status Section */}
      <Card>
        <CardHeader>
          <CardTitle>Current Status</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="text-blue-600 border-blue-600">
              Free Trial
            </Badge>
            <span className="text-sm text-muted-foreground">6 days remaining</span>
          </div>
        </CardContent>
      </Card>

      {/* Pro Plan - Billplz & FPX Payment Section */}
      <div>
        <h3 className="text-xl font-semibold mb-4">Pro Plan - Billplz & FPX Payment</h3>
        
        <Card className="border-2 border-blue-200">
          <CardHeader>
            <CardTitle className="text-2xl">Pro Plan</CardTitle>
            <CardDescription>Monthly subscription with unlimited calls</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="flex items-baseline gap-2">
              <span className="text-5xl font-bold">MYR 1.00</span>
              <span className="text-muted-foreground">/monthly</span>
            </div>

            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <CheckCircle className="w-5 h-5 text-green-500" />
                <span>Unlimited AI reply</span>
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle className="w-5 h-5 text-green-500" />
                <span>Priority support</span>
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle className="w-5 h-5 text-green-500" />
                <span>Advanced analytics</span>
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle className="w-5 h-5 text-green-500" />
                <span>Secure Billplz/FPX payments</span>
              </div>
            </div>

            <Button 
              className="w-full bg-blue-600 hover:bg-blue-700 text-white py-6 text-lg" 
              onClick={handleUpgradeClick}
              disabled={loading}
            >
              {loading ? (
                <>
                  <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                  Processing...
                </>
              ) : (
                <>
                  <CreditCard className="mr-2 h-5 w-5" />
                  Upgrade Package With Billplz
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Order History Section */}
      <Card>
        <CardHeader>
          <CardTitle>Order History</CardTitle>
          <CardDescription>
            Complete list of all your orders
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loadingOrders ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : orders.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No orders found
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Order ID</TableHead>
                  <TableHead>Product</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead>Action</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {orders.map((order) => (
                  <TableRow key={order.id}>
                    <TableCell className="font-medium">#{order.id}</TableCell>
                    <TableCell>{order.product}</TableCell>
                    <TableCell>MYR {order.amount.toFixed(2)}</TableCell>
                    <TableCell>{getStatusBadge(order.status)}</TableCell>
                    <TableCell>{new Date(order.created_at).toLocaleDateString()}</TableCell>
                    <TableCell>
                      {order.url && order.status === 'Processing' && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => window.open(order.url, '_blank')}
                        >
                          Pay Now
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
