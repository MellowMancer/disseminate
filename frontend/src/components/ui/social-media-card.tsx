import { useState } from 'react';
import { CheckCircle2, XCircle, AlertCircle, Loader2, type LucideIcon } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { DynamicShadowWrapper } from '@/components/ui/dynamic-shadow-wrapper';
import { toast } from 'sonner';

interface SocialMediaCardProps {
  platformName: string;
  platformDescription: string;
  icon: LucideIcon;
  iconBackgroundClass: string;
  isLinked: boolean | null;
  linkEndpoint: string;
  unlinkEndpoint: string;
  onStatusChange?: () => void;
}

export function SocialMediaCard({
  platformName,
  platformDescription,
  icon: Icon,
  iconBackgroundClass,
  isLinked,
  linkEndpoint,
  unlinkEndpoint,
  onStatusChange,
}: Readonly<SocialMediaCardProps>) {
  const [isUnlinking, setIsUnlinking] = useState(false);

  const handleConnect = () => {
    globalThis.location.href = linkEndpoint;
  };

  const handleDisconnect = async () => {
    setIsUnlinking(true);
    try {
      const response = await fetch(unlinkEndpoint, {
        method: 'POST',
        credentials: 'include',
      });
      if (!response.ok) throw new Error(`Failed to unlink ${platformName}`);
      toast.success(`${platformName} account disconnected`);
      onStatusChange?.();
    } catch (error) {
      toast.error(`Failed to disconnect ${platformName} account`);
      console.error(error);
    } finally {
      setIsUnlinking(false);
    }
  };
  
  const getStatus = () => {
    if (isLinked === null) {
      return {
        icon: <Loader2 className="h-5 w-5 animate-spin" />,
        text: 'Loading status...',
        className: 'text-muted-foreground',
      };
    }
    
    if (isLinked) {
      return {
        icon: <CheckCircle2 className="h-5 w-5" />,
        text: 'Connected & Active',
        className: 'text-success',
      };
    }
    
    return {
      icon: <XCircle className="h-5 w-5" />,
      text: 'Not Connected',
      className: 'text-muted-foreground',
    };
  };

  const status = getStatus();
  const showDisconnectButton = isLinked;
  const showConnectButton = !showDisconnectButton;
  const connectButtonText = isLinked ? 'Reauthorize' : 'Connect Account';

  return (
    <DynamicShadowWrapper>
      <Card>
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className={`p-3 rounded-lg ${iconBackgroundClass}`}>
                <Icon className="h-6 w-6 text-white" />
              </div>
              <div className="space-y-1">
                <CardTitle className="text-lg">{platformName}</CardTitle>
                <CardDescription className="text-sm">{platformDescription}</CardDescription>
              </div>
            </div>
          </div>
        </CardHeader>
        
        <CardContent className="space-y-4 pt-0">
          {/* Status */}
          <div className={`flex items-center gap-3 ${status.className}`}>
            {status.icon}
            <span className="text-sm font-medium">{status.text}</span>
          </div>

          {/* Actions */}
          <div className="flex gap-3 pt-2">
            {showDisconnectButton && (
              <Button
                variant="destructive"
                size="default"
                onClick={handleDisconnect}
                disabled={isUnlinking}
                className="text-sm"
              >
                {isUnlinking ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    Disconnecting...
                  </>
                ) : (
                  'Disconnect'
                )}
              </Button>
            )}
            
            {showConnectButton && (
              <Button
                variant="default"
                size="default"
                onClick={handleConnect}
                className="text-sm"
              >
                {connectButtonText}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </DynamicShadowWrapper>
  );
}


// Example function to fetch OAuth connection state
async function getOAuthStatus(platform: string): Promise<{ linked: boolean; tokenExpired: boolean }> {
  try {
    const response = await fetch(`/api/oauth/status?platform=${platform}`, {
      credentials: 'include'
    });
    
    if (!response.ok) throw new Error('Failed to fetch status');
    
    const data = await response.json();

    // Assuming response contains { linked: boolean, tokenExpiresAt: timestamp }
    const now = Date.now();
    const tokenExpiresAt = new Date(data.tokenExpiresAt).getTime();

    return {
      linked: data.linked,
      tokenExpired: tokenExpiresAt < now
    };
  } catch (error) {
    console.error('Error fetching OAuth status:', error);
    return { linked: false, tokenExpired: false };
  }
}


