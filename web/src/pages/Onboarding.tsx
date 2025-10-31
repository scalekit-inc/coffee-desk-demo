import { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { ArrowRight } from "lucide-react";
import { useForm } from "react-hook-form";
import { config } from "@/config";

interface OnboardingFormData {
  firstName: string;
  lastName: string;
  workspaceName: string;
}

const Onboarding = () => {
  const [isLoading, setIsLoading] = useState(false);

  const form = useForm<OnboardingFormData>({
    defaultValues: {
      firstName: '',
      lastName: '',
      workspaceName: ''
    }
  });

  const onSubmit = async (data: OnboardingFormData) => {
    setIsLoading(true);
    
    try {
      
      // Single onboarding API call
      const response = await fetch(`${config.backendUrl}/api/workspace/onboarding`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          workspaceName: data.workspaceName,
          firstName: data.firstName,
          lastName: data.lastName
        })
      });

      if (!response.ok) {
        throw new Error(`Onboarding failed: ${response.status}`);
      }

      
      // Redirect to /api/authorize
      window.location.href = `${config.backendUrl}/api/authorize`;
      
    } catch (error) {
      console.error("Onboarding error:", error);
      // You can add toast notification here for error handling
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Welcome to DevRamp 👋</CardTitle>
          <CardDescription>
            What should we call you—and your organization?
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="firstName"
                  rules={{ required: "First name is required" }}
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>First name</FormLabel>
                      <FormControl>
                        <Input placeholder="e.g. John" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                
                <FormField
                  control={form.control}
                  name="lastName"
                  rules={{ required: "Last name is required" }}
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Last name</FormLabel>
                      <FormControl>
                        <Input placeholder="e.g. Smith" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <FormField
                control={form.control}
                name="workspaceName"
                rules={{ required: "Organization name is required" }}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Organization name</FormLabel>
                    <FormControl>
                      <Input placeholder="e.g. Acme corp" {...field} />
                    </FormControl>
                    <FormMessage />
                    <p className="text-sm text-muted-foreground">
                      The name of your company
                    </p>
                  </FormItem>
                )}
              />

              <Button 
                type="submit" 
                className="w-full mt-6" 
                disabled={isLoading}
              >
                {isLoading ? (
                  "Creating Organization..."
                ) : (
                  <>
                    Create Organization
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </>
                )}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
};

export default Onboarding;
