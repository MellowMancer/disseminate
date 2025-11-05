import React, { useEffect, useState, useCallback } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Twitter, Instagram } from 'lucide-react';
import { SocialMediaCard } from '@/components/ui/social-media-card';
import { set } from 'react-hook-form';

const Profile: React.FC = () => {
    const [twitterLinked, setTwitterLinked] = useState<boolean | null>(null);
    const [instagramLinked, setInstagramLinked] = useState<boolean | null>(null);
    const [blueskyLinked, setBlueskyLinked] = useState<boolean | null>(null);

    useEffect(() => {
    async function fetchLinkStatus() {
      try {
        const response = await fetch('/auth/oauth_status', {
          credentials: 'include',
        });

        if (!response.ok) {
          throw new Error('Failed to fetch linking status');
        }
        const data = await response.json();

        setTwitterLinked(data.twitter_linked);
        setInstagramLinked(data.instagram_linked);
        setBlueskyLinked(data.bluesky_linked);
      } catch (error) {
        console.error('Error fetching connect status:', error);
        toast.error('Error fetching connect status');
        setTwitterLinked(false);
        setInstagramLinked(false);
        setBlueskyLinked(false);
      }
    }

    fetchLinkStatus();
  }, []);

    return (
        <div className="w-full max-w-4xl mx-auto space-y-8 pt-0 md:pt-12">
            <div className="space-y-3">
                <h1 className="text-3xl md:text-4xl font-bold text-primary">Profile & Connections</h1>
                <p className="text-base text-muted-foreground">Manage your social media accounts and posting credentials</p>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Twitter Card */}
                <SocialMediaCard
                    platformName="Twitter / X"
                    platformDescription="Connect your Twitter account"
                    icon={Twitter}
                    iconBackgroundClass="bg-twitter"
                    isLinked={twitterLinked}
                    linkEndpoint="/api/twitter/link/begin"
                    unlinkEndpoint="/api/twitter/unlink"
                    onStatusChange={() => {
                        setTwitterLinked(false);
                    }}
                />

                {/* Instagram Card */}
                <SocialMediaCard
                    platformName="Instagram"
                    platformDescription="Connect your Instagram Business account"
                    icon={Instagram}
                    iconBackgroundClass="bg-gradient-to-br from-[var(--color-instagram-from)] via-[var(--color-instagram-via)] to-[var(--color-instagram-to)]"
                    isLinked={instagramLinked}
                    linkEndpoint="/api/instagram/link/begin"
                    unlinkEndpoint="/api/instagram/unlink"
                    onStatusChange={() => {
                        setInstagramLinked(false);
                    }}
                />

                {/* Bluesky Card */}
                <SocialMediaCard
                    platformName="Bluesky"
                    platformDescription="Connect your Bluesky account"
                    icon={Twitter} // Replace with Bluesky icon when available
                    iconBackgroundClass="bg-blue-500"
                    isLinked={blueskyLinked}
                    linkEndpoint="/api/bluesky/link/begin"
                    unlinkEndpoint="/api/bluesky/unlink"
                    onStatusChange={() => {
                        setBlueskyLinked(false);
                    }}
                />  
            </div>
        </div>
    );
};

export default Profile;