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

interface UserProfile {
  id: number;
  email: string;
  full_name: string;
  gmail: string;
  phone: string;
  status: string;
  expired: string;
  is_active: boolean;
}

export default function Billings() {
  const [loading, setLoading] = useState(false);
  const [orders, setOrders] = useState<Order[]>([]);
  const [loadingOrders, setLoadingOrders] = useState(true);
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const [loadingProfile, setLoadingProfile] = useState(true);

  // Fetch user profile
  useEffect(() => {
    fetchUserProfile();
  }, []);

  // Fetch user's orders
  useEffect(() => {
    fetchOrders();
  }, []);

  const fetchUserProfile = async () => {
    try {
      // Backend uses cookie-based auth (session_token), no need for Authorization header
      const response = await fetch('/api/profile/', {
        credentials: 'include' // Important: send cookies with request
      });

      if (response.ok) {
        const result = await response.json();
        if (result.success && result.data) {
          setUserProfile(result.data);
        }
      } else {
        console.error('Failed to fetch profile:', response.statusText);
      }
    } catch (error) {
      console.error('Failed to fetch profile:', error);
    } finally {
      setLoadingProfile(false);
    }
  };

  const fetchOrders = async () => {
    try {
      // Backend uses cookie-based auth (session_token), no need for Authorization header
      const response = await fetch('/api/billing/orders?limit=50&offset=0', {
        credentials: 'include' // Important: send cookies with request
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

  const getDaysRemaining = () => {
    if (!userProfile || !userProfile.expired) return 0;
    
    const expiredDate = new Date(userProfile.expired);
    const today = new Date();
    const diffTime = expiredDate.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    
    return Math.max(0, diffDays);
  };

  const handleUpgradeClick = async () => {
    setLoading(true);

    try {
      // Backend uses cookie-based auth (session_token), send credentials with request
      const response = await fetch('/api/billing/create-order', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include', // Important: send cookies with request
        body: JSON.stringify({
          amount: 1.00,
          product: 'Pro Plan - Monthly Subscription'
        }),
      });

      const data = await response.json();

      if (response.ok) {
        if (data.payment_url) {
          toast.success('Order created! Opening payment page...');
          // Open Billplz payment page in new tab
          window.open(data.payment_url, '_blank');
          // Refresh orders list
          fetchOrders();
        } else {
          toast.success('Order created successfully!');
          fetchOrders(); // Refresh orders list
        }
      } else {
        // Check if profile is incomplete
        if (data.error === 'profile_incomplete') {
          const result = await Swal.fire({
            icon: 'warning',
            title: 'Profile Incomplete',
            html: data.message || 'Please update your profile with email, phone number, and full name before upgrading.',
            confirmButtonText: 'Go to Profile',
            confirmButtonColor: '#3085d6',
            showCancelButton: true,
            cancelButtonText: 'Cancel'
          });
          
          if (result.isConfirmed) {
            window.location.href = '/profile';
          }
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

  const getStatusBadgeVariant = (status: string) => {
    if (status === 'Trial' || status === 'Free Trial') {
      return 'outline';
    } else if (status === 'Active' || status === 'Premium') {
      return 'default';
    }
    return 'secondary';
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

      {/* Current Status Section - Dynamic */}
      <Card>
        <CardHeader>
          <CardTitle>Current Status</CardTitle>
        </CardHeader>
        <CardContent>
          {loadingProfile ? (
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              <span className="text-sm text-muted-foreground">Loading status...</span>
            </div>
          ) : userProfile ? (
            <div className="flex items-center gap-2">
              <Badge 
                variant={getStatusBadgeVariant(userProfile.status)} 
                className="text-blue-600 border-blue-600"
              >
                {userProfile.status}
              </Badge>
              <span className="text-sm text-muted-foreground">
                {getDaysRemaining()} days remaining
              </span>
            </div>
          ) : (
            <div className="text-sm text-muted-foreground">Unable to load status</div>
          )}
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
                      {order.url && order.status === 'Success' && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => window.open(order.url, '_blank')}
                        >
                          Invoice
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
