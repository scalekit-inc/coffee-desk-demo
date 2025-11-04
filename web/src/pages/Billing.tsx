import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { RBACGuard } from "@/components/RBACGuard";
import { AppSidebar } from "@/components/AppSidebar";
import { WorkspaceDropdown } from "@/components/WorkspaceDropdown";
import { 
  CreditCard, 
  Download, 
  Calendar, 
  DollarSign, 
  Users, 
  Zap,
  CheckCircle,
  AlertCircle,
  TrendingUp
} from "lucide-react";

// Mock billing data
const mockBillingData = {
  currentPlan: {
    name: "Professional",
    price: 29,
    period: "month",
    features: [
      "Up to 50 team members",
      "Advanced analytics",
      "Priority support",
      "Custom integrations",
      "API access"
    ],
    usage: {
      members: 23,
      memberLimit: 50,
      apiCalls: 125000,
      apiLimit: 200000
    }
  },
  billingHistory: [
    {
      id: "inv_001",
      date: "2024-01-15",
      amount: 29.00,
      status: "paid",
      description: "Professional Plan - January 2024"
    },
    {
      id: "inv_002", 
      date: "2023-12-15",
      amount: 29.00,
      status: "paid",
      description: "Professional Plan - December 2023"
    },
    {
      id: "inv_003",
      date: "2023-11-15", 
      amount: 29.00,
      status: "paid",
      description: "Professional Plan - November 2023"
    }
  ],
  upcomingInvoice: {
    date: "2024-02-15",
    amount: 29.00,
    description: "Professional Plan - February 2024"
  },
  paymentMethod: {
    type: "card",
    last4: "4242",
    brand: "Visa",
    expiryMonth: 12,
    expiryYear: 2025
  }
};

export default function Billing() {
  const [selectedPeriod, setSelectedPeriod] = useState("month");

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'paid':
        return <Badge variant="default" className="bg-green-100 text-green-800"><CheckCircle className="w-3 h-3 mr-1" />Paid</Badge>;
      case 'pending':
        return <Badge variant="secondary"><AlertCircle className="w-3 h-3 mr-1" />Pending</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  return (
    <RBACGuard requirePermission={["organization:settings", "workspace:billing"]}>
      <div className="flex flex-col min-h-screen w-full">
        {/* Top section with logo and organization dropdown */}
        <div className="flex items-center justify-between p-4 border-b bg-background z-10">
          <div className="flex items-center gap-4">
            {/* App logo in top-left */}
            <img 
              src="/uploads/coffee-desk-name-icon.png" 
              alt="Coffeedesk Logo"
              className="h-8 w-auto"
            />
          </div>
          
          {/* Organization dropdown in top-right */}
          <WorkspaceDropdown />
        </div>

        {/* Main content area with sidebar */}
        <div className="flex flex-1">
          {/* Left sidebar */}
          <div className="w-64 border-r bg-background">
            <AppSidebar />
          </div>
          
          {/* Main content */}
          <div className="flex-1 space-y-6 p-6">
            <div>
              <h1 className="text-3xl font-bold tracking-tight">Billing & Usage</h1>
              <p className="text-muted-foreground">
                Manage your subscription, view usage, and download invoices
              </p>
            </div>

        {/* Current Plan Overview */}
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Current Plan</CardTitle>
              <Zap className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{mockBillingData.currentPlan.name}</div>
              <p className="text-xs text-muted-foreground">
                {formatCurrency(mockBillingData.currentPlan.price)}/{mockBillingData.currentPlan.period}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Team Members</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {mockBillingData.currentPlan.usage.members}/{mockBillingData.currentPlan.usage.limit}
              </div>
              <p className="text-xs text-muted-foreground">
                {Math.round((mockBillingData.currentPlan.usage.members / mockBillingData.currentPlan.usage.limit) * 100)}% of limit used
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">API Calls</CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {mockBillingData.currentPlan.usage.apiCalls.toLocaleString()}
              </div>
              <p className="text-xs text-muted-foreground">
                {Math.round((mockBillingData.currentPlan.usage.apiCalls / 200000) * 100)}% of monthly limit
              </p>
            </CardContent>
          </Card>
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          {/* Current Plan Details */}
          <Card>
            <CardHeader>
              <CardTitle>Current Plan Details</CardTitle>
              <CardDescription>
                Your active subscription and included features
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-semibold">{mockBillingData.currentPlan.name} Plan</h3>
                  <p className="text-sm text-muted-foreground">
                    {formatCurrency(mockBillingData.currentPlan.price)} per {mockBillingData.currentPlan.period}
                  </p>
                </div>
                <Badge variant="outline">Active</Badge>
              </div>
              
              <Separator />
              
              <div>
                <h4 className="font-medium mb-2">Included Features:</h4>
                <ul className="space-y-1">
                  {mockBillingData.currentPlan.features.map((feature, index) => (
                    <li key={index} className="flex items-center text-sm">
                      <CheckCircle className="w-4 h-4 text-green-500 mr-2" />
                      {feature}
                    </li>
                  ))}
                </ul>
              </div>

              <div className="flex gap-2 pt-4">
                <Button variant="outline" size="sm">
                  Change Plan
                </Button>
                <Button variant="outline" size="sm">
                  Cancel Subscription
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Payment Method */}
          <Card>
            <CardHeader>
              <CardTitle>Payment Method</CardTitle>
              <CardDescription>
                Your current payment method on file
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <CreditCard className="h-8 w-8 text-muted-foreground" />
                  <div>
                    <p className="font-medium">
                      {mockBillingData.paymentMethod.brand} •••• {mockBillingData.paymentMethod.last4}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      Expires {mockBillingData.paymentMethod.expiryMonth}/{mockBillingData.paymentMethod.expiryYear}
                    </p>
                  </div>
                </div>
                <Badge variant="outline">Default</Badge>
              </div>
              
              <Button variant="outline" size="sm" className="w-full">
                Update Payment Method
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Upcoming Invoice */}
        <Card>
          <CardHeader>
            <CardTitle>Upcoming Invoice</CardTitle>
            <CardDescription>
              Your next billing date and amount
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <Calendar className="h-5 w-5 text-muted-foreground" />
                <div>
                  <p className="font-medium">{formatDate(mockBillingData.upcomingInvoice.date)}</p>
                  <p className="text-sm text-muted-foreground">{mockBillingData.upcomingInvoice.description}</p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-lg font-semibold">{formatCurrency(mockBillingData.upcomingInvoice.amount)}</p>
                <p className="text-sm text-muted-foreground">Auto-charged</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Billing History */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Billing History</CardTitle>
                <CardDescription>
                  Your recent invoices and payments
                </CardDescription>
              </div>
              <Button variant="outline" size="sm">
                <Download className="w-4 h-4 mr-2" />
                Download All
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {mockBillingData.billingHistory.map((invoice) => (
                <div key={invoice.id} className="flex items-center justify-between py-3 border-b last:border-b-0">
                  <div className="flex items-center space-x-4">
                    <div>
                      <p className="font-medium">{invoice.description}</p>
                      <p className="text-sm text-muted-foreground">
                        Invoice #{invoice.id} • {formatDate(invoice.date)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-4">
                    <div className="text-right">
                      <p className="font-semibold">{formatCurrency(invoice.amount)}</p>
                    </div>
                    <div className="flex items-center space-x-2">
                      {getStatusBadge(invoice.status)}
                      <Button variant="ghost" size="sm">
                        <Download className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
          </div>
        </div>
      </div>
    </RBACGuard>
  );
}
