import React, { useState } from 'react';
import { toast } from 'sonner';
import type { FormDataState } from '@/types/forms';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { TwitterTab } from '@/components/forms/TwitterTab';
import { YouTubeTab } from '@/components/forms/YoutubeTab';
import { useLocation } from 'react-router-dom';
import Carousel from "@/components/ui/carousel";

// type FormDataState = {
//   importData: {}; // TODO: Stricter Restrictions
//   twitter: { apiKey: string; apiSecret: string; accessToken: string; accessSecret: string; content: string };
//   youtubeShort: { clientId: string; clientSecret: string; title: string; description: string };
//   youtubeVideo: { clientId: string; clientSecret: string; title: string; description: string };
//   instagramPost: { username: string; password: string; caption: string };
//   instagramReel: { username: string; password: string; caption: string };
//   reddit: { clientId: string; clientSecret: string; username: string; password: string; subreddit: string; title: string; content: string };
//   mastodon: { instanceUrl: string; accessToken: string; content: string };
//   artstation: { username: string; password: string; title: string; description: string };
// };


const initialFormData: FormDataState = {
    importData: {},
    twitter: { apiKey: '', apiSecret: '', accessToken: '', accessSecret: '', content: '' },
    youtubeShort: { clientId: '', clientSecret: '', title: '', description: '' },
    youtubeVideo: { clientId: '', clientSecret: '', title: '', description: '' },
    instagramPost: { username: '', password: '', caption: '' },
    instagramReel: { username: '', password: '', caption: '' },
    reddit: { clientId: '', clientSecret: '', username: '', password: '', subreddit: '', title: '', content: '' },
    mastodon: { instanceUrl: '', accessToken: '', content: '' },
    artstation: { username: '', password: '', title: '', description: '' },
};

export function SchedulerPage() {
    const [formData, setFormData] = useState<FormDataState>(initialFormData);
    const location = useLocation();
    // const [api, setApi] = React.useState<CarouselApi>()
    const files = location.state?.files as FileList;

    type MediaItemType = {
        type: "image" | "video";
        src: string;
    };

    const mediaItems: MediaItemType[] = Array.from(files).map((file) => ({
        type: file.type.startsWith("video") ? "video" : "image", // explicitly typed
        src: URL.createObjectURL(file),
    }));

    const handleChange = <P extends keyof FormDataState>(
        platform: P,
        field: keyof FormDataState[P]
    ) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        setFormData(prev => ({
            ...prev,
            [platform]: {
                ...prev[platform],
                [field]: event.target.value,
            },
        }));
    };

    const handleSubmit = (event: React.FormEvent) => {
        event.preventDefault();
        console.log('Form Submitted:', formData);
        toast.info('Form submitted! Check the console for the data.');
    };

    return (
        <div className='w-full h-full grid place-items-center'>
            <div className="grid grid-cols-11 grid-rows-min gap-y-8 md:gap-x-8 w-full h-full content-center">
                <Card className="col-span-11 md:col-span-6 h-min ">
                    <CardHeader>
                        <CardTitle>Create a New Post</CardTitle>
                        <CardDescription>Fill out the details for each platform you want to post to.</CardDescription>
                    </CardHeader>

                    <CardContent>
                        <form onSubmit={handleSubmit}>
                            <Tabs defaultValue="twitter" className="w-full">
                                <div className="overflow-x-auto pb-2">
                                    <TabsList>
                                        <TabsList>
                                            <TabsTrigger value="twitter">Twitter / X</TabsTrigger>
                                            <TabsTrigger value="youtube">YouTube</TabsTrigger>
                                            <TabsTrigger value="twitter">Instagram</TabsTrigger>
                                            <TabsTrigger value="youtube">Reddit</TabsTrigger>
                                            <TabsTrigger value="twitter">Bluesky</TabsTrigger>
                                            <TabsTrigger value="youtube">Mastodon</TabsTrigger>
                                            <TabsTrigger value="twitter">Artstation</TabsTrigger>
                                        </TabsList>
                                    </TabsList>
                                </div>

                                {/* Render the modular components */}
                                <TabsContent value="twitter">
                                    <TwitterTab data={formData.twitter} handleChange={handleChange} />
                                </TabsContent>

                                <TabsContent value="youtube_short">
                                    <YouTubeTab data={formData.youtubeShort} handleChange={handleChange} platform="youtubeShort" />
                                </TabsContent>

                                <TabsContent value="youtube_video">
                                    <YouTubeTab data={formData.youtubeVideo} handleChange={handleChange} platform="youtubeVideo" />
                                </TabsContent>

                                {/* ... other TabsContent sections ... */}
                            </Tabs>
                            <CardFooter className="mt-8 p-0">
                                <Button type="submit" className="w-full bg-button text-white">Schedule Post</Button>
                            </CardFooter>
                        </form>
                    </CardContent>
                </Card>
                <div className='col-span-11 md:col-span-5 place-self-center w-full'>
                <Carousel mediaItems={mediaItems} />
                </div>
            </div>
        </div>
    );
}