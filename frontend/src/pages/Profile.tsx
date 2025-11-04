import React, { useEffect, useState, useCallback } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Twitter, Instagram } from 'lucide-react';
import { SocialMediaCard } from '@/components/ui/social-media-card';

const Profile: React.FC = () => {
    const [twitterLinked, setTwitterLinked] = useState<boolean | null>(null);
    const [instagramLinked, setInstagramLinked] = useState<boolean | null>(null);
    const [twitterTokenValid, setTwitterTokenValid] = useState<boolean | null>(null);
    const [instagramTokenValid, setInstagramTokenValid] = useState<boolean | null>(null);
    
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();

    const fetchTwitterStatus = useCallback(async () => {
        try {
            const response = await fetch('/api/twitter/check', {
                credentials: 'include',
            });
            if (!response.ok) throw new Error('Network response was not ok');
            const data = await response.json();
            setTwitterLinked(data.twitterLinked);
            setTwitterTokenValid(data.twitterTokenValid);
        } catch (error) {
            console.error("Failed to fetch Twitter status:", error);
            setTwitterLinked(false);
            setTwitterTokenValid(false);
        }
    }, []);

    const fetchInstagramStatus = useCallback(async () => {
        try {
            const response = await fetch('/api/instagram/check', {
                credentials: 'include',
            });
            if (!response.ok) throw new Error('Network response was not ok');
            const data = await response.json();
            setInstagramLinked(data.instagramLinked);
            setInstagramTokenValid(data.instagramTokenValid);
        } catch (error) {
            console.error("Failed to fetch Instagram status:", error);
            setInstagramLinked(false);
            setInstagramTokenValid(false);
        }
    }, []);

    useEffect(() => {
        const status = searchParams.get('status');
        const provider = searchParams.get('provider');

        if (status && provider === 'twitter') {
            if (status === 'success') {
                toast.success('Successfully connected your account!');
                fetchTwitterStatus(); 
            } else if (status === 'denied') {
                toast.error('Authorization was denied');
            } else if (status === 'error') {
                const code = searchParams.get('code') || 'Unknown error';
                toast.error(`Failed to connect: ${code}`);
            }
            
            navigate('/profile', { replace: true });
        }
        else {
            fetchTwitterStatus();
        }
    }, [fetchTwitterStatus, navigate, searchParams]);

    

    useEffect(() => {
        fetchInstagramStatus();
    }, [fetchInstagramStatus]);



    return (
        <div className="w-full max-w-4xl mx-auto space-y-8">
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
                    isTokenValid={twitterTokenValid}
                    linkEndpoint="/api/twitter/link/begin"
                    unlinkEndpoint="/api/twitter/unlink"
                    onStatusChange={() => {
                        setTwitterLinked(false);
                        setTwitterTokenValid(false);
                    }}
                />

                {/* Instagram Card */}
                <SocialMediaCard
                    platformName="Instagram"
                    platformDescription="Connect your Instagram Business account"
                    icon={Instagram}
                    iconBackgroundClass="bg-gradient-to-br from-[var(--color-instagram-from)] via-[var(--color-instagram-via)] to-[var(--color-instagram-to)]"
                    isLinked={instagramLinked}
                    isTokenValid={instagramTokenValid}
                    linkEndpoint="/api/instagram/link/begin"
                    unlinkEndpoint="/api/instagram/unlink"
                    onStatusChange={() => {
                        setInstagramLinked(false);
                        setInstagramTokenValid(false);
                    }}
                />
            </div>
        </div>
    );
};

export default Profile;