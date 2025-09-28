import React from 'react';

const Profile: React.FC = () => {
    const handleTwitterLogin = () => {
        window.location.href = 'http://localhost:8080/twitter/login';
    };

    return (
        <div>
            <h1>Profile Page</h1>
            <p>Welcome to your profile.</p>
            <button onClick={handleTwitterLogin}>Connect Twitter</button>
        </div>
    );
};

export default Profile;
