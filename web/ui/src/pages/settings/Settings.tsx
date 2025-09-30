import { useState, useEffect } from "react";
import { Shield, Key, CheckCircle, AlertCircle, Info } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";

interface SecurityStatus {
  hasPassword: boolean;
  passwordStrength?: 'weak' | 'medium' | 'strong';
}

const SettingsPage = () => {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [securityStatus, setSecurityStatus] = useState<SecurityStatus>({ hasPassword: false });
  const [isLoading, setIsLoading] = useState(false);
  const { toast } = useToast();

  // Check if server requires password (via API call)
  useEffect(() => {
    const checkSecurityStatus = async () => {
      try {
        const response = await fetch('/api/security/status');
        if (response.ok) {
          const data = await response.json();
          setSecurityStatus({ 
            hasPassword: data.password_enabled, 
            passwordStrength: data.password_enabled ? 'medium' : undefined 
          });
        }
      } catch (error) {
        // Fallback to cookie check
        const hasAuthCookie = document.cookie.includes('gowitness_auth');
        setSecurityStatus({ hasPassword: hasAuthCookie, passwordStrength: hasAuthCookie ? 'medium' : undefined });
      }
    };

    checkSecurityStatus();
  }, []);

  const evaluatePasswordStrength = (password: string): 'weak' | 'medium' | 'strong' => {
    if (password.length < 8) return 'weak';
    if (password.length >= 12 && /(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])/.test(password)) {
      return 'strong';
    }
    return 'medium';
  };

  const handlePasswordUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      // Validation
      if (newPassword !== confirmPassword) {
        toast({
          title: "Password Error",
          description: "New passwords don't match",
          variant: "destructive",
        });
        return;
      }

      if (newPassword.length < 8) {
        toast({
          title: "Password Error",
          description: "Password must be at least 8 characters long",
          variant: "destructive",
        });
        return;
      }

      // For now, we'll show a message about server restart
      // In a real implementation, this would make an API call
      toast({
        title: "Password Configuration",
        description: "To enable password protection, restart the gowitness server with the --password flag",
        variant: "default",
      });

      // Clear form
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");

    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update password settings",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const getStrengthColor = (strength: string) => {
    switch (strength) {
      case 'weak': return 'bg-red-500';
      case 'medium': return 'bg-yellow-500';
      case 'strong': return 'bg-green-500';
      default: return 'bg-gray-500';
    }
  };

  const getStrengthBadgeVariant = (strength: string) => {
    switch (strength) {
      case 'weak': return 'destructive' as const;
      case 'medium': return 'secondary' as const;
      case 'strong': return 'default' as const;
      default: return 'secondary' as const;
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground">
          Configure security and authentication settings for your gowitness server.
        </p>
      </div>

      {/* Security Status Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Security Status
          </CardTitle>
          <CardDescription>
            Current security configuration for your gowitness server
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Key className="h-4 w-4" />
              <span>Password Protection</span>
            </div>
            <div className="flex items-center gap-2">
              {securityStatus.hasPassword ? (
                <>
                  <CheckCircle className="h-4 w-4 text-green-500" />
                  <Badge variant="default">Enabled</Badge>
                </>
              ) : (
                <>
                  <AlertCircle className="h-4 w-4 text-red-500" />
                  <Badge variant="destructive">Disabled</Badge>
                </>
              )}
            </div>
          </div>

          {securityStatus.hasPassword && securityStatus.passwordStrength && (
            <div className="flex items-center justify-between">
              <span>Password Strength</span>
              <Badge variant={getStrengthBadgeVariant(securityStatus.passwordStrength)}>
                {securityStatus.passwordStrength.charAt(0).toUpperCase() + securityStatus.passwordStrength.slice(1)}
              </Badge>
            </div>
          )}

          {!securityStatus.hasPassword && (
            <Card className="border-orange-200 bg-orange-50 dark:border-orange-800 dark:bg-orange-950">
              <CardContent className="flex gap-2 p-4">
                <Info className="h-4 w-4 text-orange-600 mt-0.5 flex-shrink-0" />
                <div className="text-sm text-orange-800 dark:text-orange-200">
                  Password protection is currently disabled. Anyone with access to this server can view all data.
                  To enable password protection, restart the server with the <code className="bg-orange-200 dark:bg-orange-800 px-1 py-0.5 rounded text-xs">--password</code> flag.
                </div>
              </CardContent>
            </Card>
          )}
        </CardContent>
      </Card>

      {/* Password Configuration Card */}
      <Card>
        <CardHeader>
          <CardTitle>Password Configuration</CardTitle>
          <CardDescription>
            {securityStatus.hasPassword 
              ? "Update your server password" 
              : "Configure password protection for your server"
            }
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Card className="mb-6 border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-950">
            <CardContent className="flex gap-2 p-4">
              <Info className="h-4 w-4 text-blue-600 mt-0.5 flex-shrink-0" />
              <div className="text-sm text-blue-800 dark:text-blue-200">
                <strong>Note:</strong> Password changes require a server restart with the new password.
                Use the command: <code className="bg-blue-200 dark:bg-blue-800 px-1 py-0.5 rounded text-xs">gowitness report server --password "your-new-password"</code>
              </div>
            </CardContent>
          </Card>

          <form onSubmit={handlePasswordUpdate} className="space-y-4">
            {securityStatus.hasPassword && (
              <div className="space-y-2">
                <Label htmlFor="current-password">Current Password</Label>
                <Input
                  id="current-password"
                  type="password"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                  placeholder="Enter current password"
                />
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="new-password">
                {securityStatus.hasPassword ? "New Password" : "Password"}
              </Label>
              <Input
                id="new-password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="Enter new password"
              />
              {newPassword && (
                <div className="flex items-center gap-2 text-sm">
                  <span>Strength:</span>
                  <div className={`h-2 w-16 rounded ${getStrengthColor(evaluatePasswordStrength(newPassword))}`} />
                  <Badge variant={getStrengthBadgeVariant(evaluatePasswordStrength(newPassword))} className="text-xs">
                    {evaluatePasswordStrength(newPassword).charAt(0).toUpperCase() + evaluatePasswordStrength(newPassword).slice(1)}
                  </Badge>
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="confirm-password">Confirm Password</Label>
              <Input
                id="confirm-password"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Confirm new password"
              />
            </div>

            <div className="border-t pt-4 mt-6">
              <Button 
                type="submit" 
                disabled={isLoading || !newPassword || newPassword !== confirmPassword}
                className="flex items-center gap-2"
              >
                <Key className="h-4 w-4" />
                {isLoading ? "Updating..." : securityStatus.hasPassword ? "Update Password" : "Generate Configuration"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {/* Security Recommendations */}
      <Card>
        <CardHeader>
          <CardTitle>Security Recommendations</CardTitle>
          <CardDescription>
            Best practices for securing your gowitness server
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-start gap-3">
            <CheckCircle className="h-5 w-5 text-green-500 mt-0.5 flex-shrink-0" />
            <div>
              <h4 className="font-medium">Use Strong Passwords</h4>
              <p className="text-sm text-muted-foreground">
                Use at least 12 characters with a mix of uppercase, lowercase, numbers, and symbols.
              </p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <CheckCircle className="h-5 w-5 text-green-500 mt-0.5 flex-shrink-0" />
            <div>
              <h4 className="font-medium">Enable HTTPS</h4>
              <p className="text-sm text-muted-foreground">
                Run the server behind a reverse proxy with SSL/TLS encryption in production.
              </p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <CheckCircle className="h-5 w-5 text-green-500 mt-0.5 flex-shrink-0" />
            <div>
              <h4 className="font-medium">Restrict Network Access</h4>
              <p className="text-sm text-muted-foreground">
                Bind to specific interfaces and use firewalls to control access to the server.
              </p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <CheckCircle className="h-5 w-5 text-green-500 mt-0.5 flex-shrink-0" />
            <div>
              <h4 className="font-medium">Regular Updates</h4>
              <p className="text-sm text-muted-foreground">
                Keep gowitness updated to the latest version for security patches.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default SettingsPage;
