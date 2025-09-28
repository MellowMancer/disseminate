import React, { useEffect, useState, useCallback } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { toast } from 'sonner'; 

const Profile: React.FC = () => {
    const [twitterLinked, setTwitterLinked] = useState<boolean | null>(null);
    const [tokenValid, setTokenValid] = useState<boolean | null>(null);
    
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
            setTokenValid(data.tokenValid);
        } catch (error) {
            console.error("Failed to fetch Twitter status:", error);
            setTwitterLinked(false);
            setTokenValid(false);
        }
    }, []);

    useEffect(() => {
        const status = searchParams.get('status');
        const provider = searchParams.get('provider');

        if (status && provider === 'twitter') {
            // The toast function calls remain exactly the same
            if (status === 'success') {
                toast.success('Successfully connected your Twitter account!');
                fetchTwitterStatus(); 
            } else if (status === 'denied') {
                toast.error('Authorization was denied for Twitter.');
            } else if (status === 'error') {
                const code = searchParams.get('code') || 'Unknown error';
                toast.error(`Failed to connect Twitter: ${code}`);
            }
            
            navigate('/profile', { replace: true });
        } else {
            fetchTwitterStatus();
        }
    }, [fetchTwitterStatus, navigate, searchParams]);

    const handleTwitterLogin = () => {
        window.location.href = '/api/twitter/link/begin';
    };

    const renderTwitterStatus = () => {
        if (twitterLinked === null) {
            return <p>Loading Twitter status...</p>;
        }
        if (twitterLinked && tokenValid) {
            return <p style={{ color: 'green' }}>✅ Twitter is connected and ready to post.</p>;
        }
        if (twitterLinked && !tokenValid) {
            return (
                <>
                    <p style={{ color: 'orange' }}>⚠️ Your Twitter connection has expired.</p>
                    <button onClick={handleTwitterLogin}>Reauthorize Twitter</button>
                </>
            );
        }
        return <button onClick={handleTwitterLogin}>Connect Twitter</button>;
    };

    return (
        <div>
            <h1>Profile Page</h1>
            <p>Welcome to your profile. Manage your social media connections here.</p>
            
            <hr style={{ margin: '20px 0' }} />

            <h3>Connections</h3>
            {renderTwitterStatus()}
        </div>
    );
};

export default Profile;