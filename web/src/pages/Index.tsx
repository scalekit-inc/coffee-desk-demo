import { Button } from "@/components/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ArrowRight, Users, Target, TrendingUp, Menu, X } from "lucide-react";
import { useState } from "react";
import { config } from "@/config";

const Index = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

const handleAuth = (type: 'signin' | 'signup') => {
  const promptParam = type === 'signup' ? '?prompt=create' : '';
  window.location.href = `${config.backendUrl}/api/authorize${promptParam}`;
};

  const features = [
    {
      icon: <Users className="h-8 w-8 text-primary" />,
      title: "Team Collaboration",
      description: "Seamlessly work together with your team in real-time, no matter where they are."
    },
    {
      icon: <Target className="h-8 w-8 text-primary" />,
      title: "Goal Tracking",
      description: "Set, track, and achieve your business objectives with powerful analytics."
    },
    {
      icon: <TrendingUp className="h-8 w-8 text-primary" />,
      title: "Growth Analytics",
      description: "Make data-driven decisions with comprehensive insights and reporting."
    }
  ];

  return (
    <div className="min-h-screen bg-background">
      {/* Navigation */}
      <nav className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50">
        <div className="container mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <img 
<<<<<<< Updated upstream
                src="/uploads/coffee-desk-name-icon.png" 
=======
                src="/uploads/name_icon.png" 
>>>>>>> Stashed changes
                alt="Coffeedesk Logo" 
                className="h-8 w-auto"
              />
            </div>
            
            {/* Desktop Navigation */}
            <div className="hidden md:block">
              <div className="ml-10 flex items-baseline space-x-8">
                <a href="#features" className="text-foreground hover:text-primary transition-colors">Features</a>
                <a href="/dashboard" className="text-foreground hover:text-primary transition-colors">Dashboard</a>
                <a href="#contact" className="text-foreground hover:text-primary transition-colors">Contact</a>
                <Button variant="outline" size="sm" onClick={() =>handleAuth('signin')}>Sign In</Button>
                <Button size="sm" onClick={() => handleAuth('signup')}>Get Started</Button>
              </div>
            </div>

            {/* Mobile menu button */}
            <div className="md:hidden">
              <Button 
                variant="ghost" 
                size="sm"
                onClick={() => setIsMenuOpen(!isMenuOpen)}
              >
                {isMenuOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
              </Button>
            </div>
          </div>
        </div>

        {/* Mobile Navigation */}
        {isMenuOpen && (
          <div className="md:hidden">
            <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3 border-t">
              <a href="#features" className="block px-3 py-2 text-foreground hover:text-primary">Features</a>
              <a href="/dashboard" className="block px-3 py-2 text-foreground hover:text-primary">Dashboard</a>
              <a href="#contact" className="block px-3 py-2 text-foreground hover:text-primary">Contact</a>
              <div className="px-3 py-2 space-y-2">
                <Button variant="outline" size="sm" className="w-full" onClick={() => handleAuth('signin')}>Sign In</Button>
                <Button size="sm" className="w-full" onClick={() => handleAuth('signup')}>Get Started</Button>
              </div>
            </div>
          </div>
        )}
      </nav>

      {/* Hero Section */}
      <section className="py-20 lg:py-32 bg-gradient-to-br from-background to-muted/20">
        <div className="container mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center max-w-4xl mx-auto">
            <Badge variant="secondary" className="mb-6 text-sm">
              🚀 Now trusted by 10,000+ businesses worldwide
            </Badge>
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-foreground mb-6 leading-tight">
              Streamline Your Business
              <span className="text-primary block">Operations Today</span>
            </h1>
            <p className="text-xl text-muted-foreground mb-8 max-w-2xl mx-auto leading-relaxed">
              The all-in-one platform that helps teams collaborate better, track goals effectively, 
              and grow faster with powerful analytics and insights.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-12">
              <Button size="lg" className="text-lg px-8 py-6">
                Start Free Trial
                <ArrowRight className="ml-2 h-5 w-5" />
              </Button>
              <Button variant="outline" size="lg" className="text-lg px-8 py-6">
                Watch Demo
              </Button>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8 text-center">
              <div>
                <div className="text-3xl font-bold text-primary">99.9%</div>
                <div className="text-muted-foreground">Uptime</div>
              </div>
              <div>
                <div className="text-3xl font-bold text-primary">10K+</div>
                <div className="text-muted-foreground">Happy Customers</div>
              </div>
              <div>
                <div className="text-3xl font-bold text-primary">40%</div>
                <div className="text-muted-foreground">Productivity Boost</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="py-20 bg-background">
        <div className="container mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-foreground mb-4">
              Everything You Need to Succeed
            </h2>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              Powerful features designed to help your business thrive in today's competitive landscape.
            </p>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {features.map((feature, index) => (
              <Card key={index} className="border-border hover:shadow-lg transition-shadow duration-300">
                <CardHeader className="text-center">
                  <div className="flex justify-center mb-4">
                    {feature.icon}
                  </div>
                  <CardTitle className="text-xl mb-2">{feature.title}</CardTitle>
                  <CardDescription className="text-base">
                    {feature.description}
                  </CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section id="contact" className="py-20 bg-primary text-primary-foreground">
        <div className="container mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl lg:text-4xl font-bold mb-6">
            Ready to Transform Your Business?
          </h2>
          <p className="text-xl mb-8 max-w-2xl mx-auto opacity-90">
<<<<<<< Updated upstream
            Join thousands of companies already using Coffee desk to streamline operations and drive growth.
=======
            Join thousands of companies already using Coffeedesk to streamline operations and drive growth.
>>>>>>> Stashed changes
          </p>
          <Button size="lg" variant="secondary" className="text-lg px-8 py-6">
            Start Your Free Trial Today
            <ArrowRight className="ml-2 h-5 w-5" />
          </Button>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-muted/20 py-12">
        <div className="container mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            <div>
              <img 
<<<<<<< Updated upstream
                src="/uploads/coffee-desk-name-icon.png" 
=======
                src="/uploads/name_icon.png" 
>>>>>>> Stashed changes
                alt="Coffeedesk Logo" 
                className="h-8 w-auto mb-4"
              />
              <p className="text-muted-foreground mb-4">
                Streamlining business operations for teams worldwide.
              </p>
            </div>
            <div>
              <h3 className="font-semibold text-foreground mb-4">Product</h3>
              <ul className="space-y-2 text-muted-foreground">
                <li><a href="#" className="hover:text-primary transition-colors">Features</a></li>
                <li><a href="#" className="hover:text-primary transition-colors">Integrations</a></li>
                <li><a href="#" className="hover:text-primary transition-colors">API</a></li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-foreground mb-4">Company</h3>
              <ul className="space-y-2 text-muted-foreground">
                <li><a href="#" className="hover:text-primary transition-colors">About</a></li>
                <li><a href="#" className="hover:text-primary transition-colors">Careers</a></li>
                <li><a href="#" className="hover:text-primary transition-colors">Blog</a></li>
                <li><a href="#" className="hover:text-primary transition-colors">Contact</a></li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-foreground mb-4">Support</h3>
              <ul className="space-y-2 text-muted-foreground">
                <li><a href="#" className="hover:text-primary transition-colors">Help Center</a></li>
                <li><a href="#" className="hover:text-primary transition-colors">Documentation</a></li>
                <li><a href="#" className="hover:text-primary transition-colors">Status</a></li>
                <li><a href="#" className="hover:text-primary transition-colors">Security</a></li>
              </ul>
            </div>
          </div>
          <div className="border-t border-border mt-8 pt-8 text-center text-muted-foreground">
<<<<<<< Updated upstream
            <p>&copy; 2024 Coffee Desk. All rights reserved.</p>
=======
            <p>&copy; 2024 Coffeedesk. All rights reserved.</p>
>>>>>>> Stashed changes
          </div>
        </div>
      </footer>
    </div>
  );
};

export default Index;
