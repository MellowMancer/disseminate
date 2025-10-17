import React, { useEffect, useState, useCallback } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { toast } from 'sonner'; 

const Profile: React.FC = () => {
    const [twitterLinked, setTwitterLinked] = useState<boolean | null>(null);
    const [instagramLinked, setInstagramLinked] = useState<boolean | null>(null);
    const [twitterTokenValid, setTwitterTokenValid] = useState<boolean | null>(null);
    const [instagramTokenValid, setInstagramTokenValid] = useState<boolean | null>(null);
    
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();

    const handleTwitterLogin = () => {
        window.location.href = '/api/twitter/link/begin';
    };

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



    const renderTwitterStatus = () => {
        if (twitterLinked === null) {
            return <p>Loading Twitter status...</p>;
        }
        if (twitterLinked && twitterTokenValid) {
            return <p style={{ color: 'green' }}>Twitter is connected and ready to post.</p>;
        }
        if (twitterLinked && !twitterTokenValid) {
            return (
                <>
                    <p style={{ color: 'orange' }}>Your Twitter connection has expired.</p>
                    <button onClick={handleTwitterLogin}>Reauthorize Twitter</button>
                </>
            );
        }
        return <button onClick={handleTwitterLogin}>Connect Twitter</button>;
    };

    const renderInstagramStatus = () => {
        if (instagramLinked === null) {
            return <p>Loading Instagram status...</p>;
        }
        if (instagramLinked && instagramTokenValid) {
            return <p style={{ color: 'green' }}>Instagram is connected and ready to post.</p>;
        }
        if (instagramLinked && !instagramTokenValid) {
            return (
                <>
                    <p style={{ color: 'orange' }}>Your Instagram connection has expired.</p>
                    <button onClick={() => window.location.href = '/api/instagram/link/begin'}>Reauthorize Instagram</button>
                </>
            );
        }
        return <button onClick={() => window.location.href = '/api/instagram/link/begin'}>Connect Instagram</button>;
    }

    return (
        <div>
            <h1>Profile Page</h1>
            <p>Welcome to your profile. Manage your social media connections here.</p>
            
            <hr style={{ margin: '20px 0' }} />

            <h3>Connections</h3>
            {renderTwitterStatus()}
            {renderInstagramStatus()}
        </div>
    );
};

export default Profile;