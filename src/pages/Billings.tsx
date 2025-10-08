import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from 'sonner';
import { CreditCard, Package, CheckCircle, Clock, XCircle, Loader2 } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

interface Order {
  id: number;
  customer_email: string;
  customer_name: string;
  billing_phone: string;
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
  
  const [formData, setFormData] = useState({
    customer_email: '',
    customer_name: '',
    billing_phone: '',
    billing_address: '',
    billing_city: '',
    billing_state: '',
    billing_postcode: '',
    amount: '1.00', // Set to RM 1 as requested
    product: 'NodePath Subscription',
    method: 'billplz'
  });

  // Fetch user's orders
  useEffect(() => {
    fetchOrders();
  }, []);

  const fetchOrders = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch('/api/billing/orders?limit=10&offset=0', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const data = await response.json();
        setOrders(data.orders || []);
      }
    } catch (error) {
      console.error('Failed to fetch orders:', error);
    } finally {
      setLoadingOrders(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await fetch('/api/billing/create-order', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...formData,
          amount: parseFloat(formData.amount)
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
          // Refresh orders list
          fetchOrders();
          // Reset form
          setFormData({
            ...formData,
            customer_email: '',
            customer_name: '',
            billing_phone: '',
            billing_address: '',
            billing_city: '',
            billing_state: '',
            billing_postcode: ''
          });
        }
      } else {
        toast.error(data.error || 'Failed to create order');
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
          Billings
        </h2>
        <p className="text-muted-foreground">
          Manage your payments and subscriptions
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* Create Order Form */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <CreditCard className="w-5 h-5" />
              Create New Order
            </CardTitle>
            <CardDescription>
              Fill in your details to create a payment order (RM 1.00)
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="customer_name">Full Name *</Label>
                <Input
                  id="customer_name"
                  name="customer_name"
                  value={formData.customer_name}
                  onChange={handleInputChange}
                  required
                  placeholder="John Doe"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="customer_email">Email *</Label>
                <Input
                  id="customer_email"
                  name="customer_email"
                  type="email"
                  value={formData.customer_email}
                  onChange={handleInputChange}
                  required
                  placeholder="john@example.com"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="billing_phone">Phone Number *</Label>
                <Input
                  id="billing_phone"
                  name="billing_phone"
                  value={formData.billing_phone}
                  onChange={handleInputChange}
                  required
                  placeholder="0123456789"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="billing_address">Address (Optional)</Label>
                <Input
                  id="billing_address"
                  name="billing_address"
                  value={formData.billing_address}
                  onChange={handleInputChange}
                  placeholder="123 Main Street"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="billing_city">City</Label>
                  <Input
                    id="billing_city"
                    name="billing_city"
                    value={formData.billing_city}
                    onChange={handleInputChange}
                    placeholder="Kuala Lumpur"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="billing_state">State</Label>
                  <Input
                    id="billing_state"
                    name="billing_state"
                    value={formData.billing_state}
                    onChange={handleInputChange}
                    placeholder="Selangor"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="billing_postcode">Postcode</Label>
                <Input
                  id="billing_postcode"
                  name="billing_postcode"
                  value={formData.billing_postcode}
                  onChange={handleInputChange}
                  placeholder="50000"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="product">Product *</Label>
                <Input
                  id="product"
                  name="product"
                  value={formData.product}
                  onChange={handleInputChange}
                  required
                  placeholder="NodePath Subscription"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="amount">Amount (RM) *</Label>
                <Input
                  id="amount"
                  name="amount"
                  type="number"
                  step="0.01"
                  min="1"
                  value={formData.amount}
                  onChange={handleInputChange}
                  required
                  disabled
                  className="bg-muted"
                />
                <p className="text-xs text-muted-foreground">
                  Test payment amount: RM 1.00
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="method">Payment Method *</Label>
                <select
                  id="method"
                  name="method"
                  value={formData.method}
                  onChange={(e) => setFormData(prev => ({ ...prev, method: e.target.value }))}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  required
                >
                  <option value="billplz">Online Banking (Billplz)</option>
                  <option value="cod">Cash on Delivery</option>
                </select>
              </div>

              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Processing...
                  </>
                ) : (
                  <>
                    <Package className="mr-2 h-4 w-4" />
                    Create Order (RM {formData.amount})
                  </>
                )}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Order History */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Package className="w-5 h-5" />
              Recent Orders
            </CardTitle>
            <CardDescription>
              Your recent payment orders
            </CardDescription>
          </CardHeader>
          <CardContent>
            {loadingOrders ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : orders.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No orders yet. Create your first order!
              </div>
            ) : (
              <div className="space-y-4">
                {orders.map((order) => (
                  <div key={order.id} className="border rounded-lg p-4 space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="font-medium">Order #{order.id}</div>
                      {getStatusBadge(order.status)}
                    </div>
                    <div className="text-sm text-muted-foreground space-y-1">
                      <div>{order.product}</div>
                      <div className="font-semibold text-foreground">RM {order.amount.toFixed(2)}</div>
                      <div>{order.method === 'billplz' ? 'Online Banking' : 'Cash on Delivery'}</div>
                      <div className="text-xs">{new Date(order.created_at).toLocaleString()}</div>
                    </div>
                    {order.url && order.status === 'Processing' && (
                      <Button
                        size="sm"
                        variant="outline"
                        className="w-full mt-2"
                        onClick={() => window.open(order.url, '_blank')}
                      >
                        Complete Payment
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* All Orders Table */}
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
                  <TableHead>Method</TableHead>
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
                    <TableCell>RM {order.amount.toFixed(2)}</TableCell>
                    <TableCell>
                      {order.method === 'billplz' ? 'Online Banking' : 'COD'}
                    </TableCell>
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
