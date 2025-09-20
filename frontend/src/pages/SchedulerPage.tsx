import React, { useState } from 'react';
import { toast } from 'sonner';
import type { FormDataState } from '@/types/forms';
import splatImage from '@/assets/splat_poster.png';
import ss from '@/assets/ss.png';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { TwitterTab } from '@/components/forms/TwitterTab';
import { YouTubeTab } from '@/components/forms/YoutubeTab';
import { InputWithLabel } from '@/components/ui/inputWithLabel';
import { useLocation } from 'react-router-dom';
import {
    Carousel,
    CarouselContent,
    CarouselItem,
    CarouselNext,
    CarouselPrevious,
    CarouselDots
} from "@/components/ui/carousel"
import Fade from 'embla-carousel-fade'; 


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
    const files = location.state?.files as FileList;

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

    const handleExportKeys = () => {
        const keysToExport = { twitter: { apiKey: formData.twitter.apiKey, /*...*/ } };
        const jsonString = JSON.stringify(keysToExport, null, 2);
        const blob = new Blob([jsonString], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = 'disseminate-keys.json';
        link.click();
        URL.revokeObjectURL(url);
        toast.success('Developer keys exported successfully!');
    };

    const handleSubmit = (event: React.FormEvent) => {
        event.preventDefault();
        console.log('Form Submitted:', formData);
        toast.info('Form submitted! Check the console for the data.');
    };

    return (
        <div className="min-h-screen w-screen grid grid-cols-1 lg:grid-cols-10 gap-10 lg:gap-5 bg-background py-16 px-8 lg:py-8 lg:px-12 text-white overflow-y-scroll">
            <div className="lg:col-span-6 w-full lg:px-8 relative place-self-center">
                <Card className="bg-card text-slate-800">
                    <CardHeader>
                        <CardTitle>Create a New Post</CardTitle>
                        <CardDescription>Fill out the details for each platform you want to post to.</CardDescription>
                        <div className="grid grid-cols-1 lg:grid-cols-2 gap-2 z-10">
                            <Button variant="outline" onClick={handleExportKeys}>Import Keys</Button>
                            <Button variant="outline" onClick={handleExportKeys}>Export Keys</Button>
                        </div>
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
            </div>
            <div className='lg:col-span-4 max-h-screen flex justify-center'>
                <Carousel opts={{
                    loop: true,
                }} plugins={[Fade()]}
                    className="w-3/4 place-self-center overscroll-contain">
                    <CarouselContent>
                        {Array.from(files).map((file, index) => {
                            const url = URL.createObjectURL(file);
                            if (file.type.startsWith("image")) {
                                return (
                                    <CarouselItem key={index}>
                                        <div className='flex aspect-square items-center justify-center p-2' >
                                                <img className="rounded-sm border-1 border-t-24 shadow-(--shadow-override) border-card-outline" src={url} alt={file.name} style={{ maxWidth: '100%', maxHeight: '100%' }} />
                                        </div>
                                    </CarouselItem>
                                );
                            } else if (file.type.startsWith("video")) {
                                return (
                                    <CarouselItem key={index}>
                                        <div className='flex aspect-square items-center justify-center p-2' >
                                                <video className="rounded-sm border-1 border-t-24 shadow-(--shadow-override) border-card-outline" src={url} controls style={{ maxWidth: '100%', maxHeight: '100%' }} />
                                        </div>
                                    </CarouselItem>
                                );
                            }
                            return null;
                        })}
                    </CarouselContent>

                    <CarouselPrevious />
                    <CarouselNext />
                    <CarouselDots/>
                </Carousel>
            </div>
        </div>
    );
}