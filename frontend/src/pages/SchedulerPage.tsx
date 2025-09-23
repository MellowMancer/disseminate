import React, { useState } from 'react';
import { toast } from 'sonner';
import type { FormDataState } from '@/types/forms';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { TwitterTab } from '@/pages/schedule/TwitterTab';
import { YouTubeTab } from '@/pages/schedule/YoutubeTab';
import { useLocation } from 'react-router-dom';
import Carousel from "@/components/ui/carousel";
import { Dialog } from "@/components/ui/dialog";
import { DynamicShadowWrapper } from "@/components/ui/dynamicShadowWrapper";

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
  const [activeTab, setActiveTab] = React.useState<TabKey>('twitter');

  const files = location.state?.files as FileList;

  type TabKey = 'twitter' | 'youtube' | 'instagram' | 'reddit' | 'mastodon' | 'artstation';


  type MediaItemType = {
    id: string,
    type: "image" | "video";
    src: string;
  };

  const mediaItems: MediaItemType[] = Array.from(files).map((file) => ({
    id: file.name + "_" + file.type + " " + file.lastModified,
    type: file.type.startsWith("video") ? "video" : "image", // explicitly typed
    src: URL.createObjectURL(file),
  }));

  const twitterMediaItems = appendTabToIds(mediaItems, "twitter");
  const youtubeMediaItems = appendTabToIds(mediaItems, "youtube");
  const instagramMediaItems = appendTabToIds(mediaItems, "instagram");
  const redditMediaItems = appendTabToIds(mediaItems, "reddit");
  const mastodonMediaItems = appendTabToIds(mediaItems, "mastodon");
  const artstationMediaItems = appendTabToIds(mediaItems, "artstation");

  const carouselsData: Record<TabKey, MediaItemType[]> = {
    twitter: twitterMediaItems,
    youtube: youtubeMediaItems,
    instagram: instagramMediaItems,
    reddit: redditMediaItems,
    mastodon: mastodonMediaItems,
    artstation: artstationMediaItems,
  };




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

  function appendTabToIds(items: MediaItemType[], tabName: string): MediaItemType[] {
    return items.map(item => ({
      ...item,
      id: `${tabName}_${item.id}`,
    }));
  }
  console.log('Active Tab:', activeTab);
  console.log('Carousel mediaItems for active tab:', carouselsData[activeTab]);
  return (
    <div className='w-full h-full grid place-items-center'>
      <div className="grid grid-cols-11 grid-rows-min gap-y-8 md:gap-x-8 w-full h-full content-center">
        <DynamicShadowWrapper className="col-span-11 md:col-span-6 h-min ">
        <Card >
          <CardHeader>
            <CardTitle>Create a New Post</CardTitle>
            <CardDescription>Fill out the details for each platform you want to post to.</CardDescription>
          </CardHeader>

          <CardContent>
            <form onSubmit={handleSubmit}>
              <Tabs
                value={activeTab}
                onValueChange={(value: string) => setActiveTab(value as TabKey)}
                className="w-full"
              >
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
        </DynamicShadowWrapper>
        <div className='col-span-11 md:col-span-5 place-self-center w-full'>
          {Object.entries(carouselsData).map(([tabKey, mediaItems]) => (
            <Dialog>
                <div key={tabKey} style={{ display: tabKey === activeTab ? 'block' : 'none' }}>
                  <Carousel mediaItems={mediaItems} />
                </div>
            </Dialog>
          ))}
        </div>

      </div>
    </div >
  );
}