import React, { useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { DynamicShadowWrapper } from '@/components/ui/dynamic-shadow-wrapper';
import Carousel from '@/components/ui/carousel';

import { useMediaManager } from '@/lib/useMediaManager';
import { TwitterTab } from '@/pages/schedule/TwitterTab';
import { InstagramTab } from '@/pages/schedule/InstagramTab';
import type { TabKey, FormDataState } from '@/types/types';

export function SchedulerPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<TabKey>('twitter');
  const {
    isReady,
    formData,
    setFormData,
    isSubmitting,
    handleSubmit,
    carouselMediaItems,
    selectedMedia,
    handleMediaSelectionChange,
    handleMediaUpdate,
    handleRevertMedia,
    setOrderedMediaByTab,
    mediaOverrides,
  } = useMediaManager(location.state?.files, activeTab);

  if (!isReady) {
    return (
      <div className="w-full h-full grid place-items-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Loading Media...</CardTitle>
            <CardDescription>
              If you refreshed this page, the media files were lost and you'll need to go
              back.
            </CardDescription>
          </CardHeader>
          <CardFooter>
            <Button className="w-full" onClick={() => navigate('/')}>
              Go Back to Upload
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  const handleChange = <P extends keyof FormDataState>(
    platform: P,
    field: keyof FormDataState[P],
  ) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData(prev => ({
      ...prev,
      [platform]: { ...prev[platform], [field]: event.target.value },
    }));
  };

  return (
    <div className="w-full h-full grid place-items-center p-4">
      <div className="grid grid-cols-11 grid-rows-min gap-y-8 md:gap-x-8 w-full h-full content-center max-w-6xl">
        <DynamicShadowWrapper className="col-span-11 md:col-span-6 h-min">
          <Card>
            <CardHeader>
              <CardTitle>Create a New Post</CardTitle>
              <CardDescription>Fill out the details for each platform you want to post to.</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit}>
                <Tabs
                  value={activeTab}
                  onValueChange={value => setActiveTab(value as TabKey)}
                  className="w-full"
                >
                  <div className="overflow-x-auto pb-2">
                    <TabsList>
                      <TabsTrigger value="twitter">Twitter / X</TabsTrigger>
                      <TabsTrigger value="youtube">YouTube</TabsTrigger>
                      <TabsTrigger value="instagram">Instagram</TabsTrigger>
                      <TabsTrigger value="reddit">Reddit</TabsTrigger>
                      <TabsTrigger value="mastodon">Mastodon</TabsTrigger>
                      <TabsTrigger value="artstation">Artstation</TabsTrigger>
                    </TabsList>
                  </div>

                  <TabsContent value="twitter">
                    <TwitterTab data={formData.twitter} handleChange={handleChange} />
                  </TabsContent>

                  <TabsContent value="youtube">
                    <div className="mt-4">Youtube Posting coming soon!</div>
                  </TabsContent>

                  <TabsContent value="instagram">
                    <InstagramTab data={formData.instagram} handleChange={handleChange} />
                  </TabsContent>

                  <TabsContent value="reddit">
                    <div className="mt-4">Reddit Posting coming soon!</div>
                  </TabsContent>

                  <TabsContent value="mastodon">
                    <div className="mt-4">Mastodon Posting coming soon!</div>
                  </TabsContent>

                  <TabsContent value="artstation">
                    <div className="mt-4">Artstation Posting coming soon!</div>
                  </TabsContent>
                </Tabs>

                <CardFooter className="mt-8 p-0">
                  <Button
                    type="submit"
                    className="w-full bg-primary text-primary-foreground"
                    disabled={isSubmitting}
                  >
                    {isSubmitting ? 'Posting...' : `Post for ${activeTab}`}
                  </Button>
                </CardFooter>
              </form>
            </CardContent>
          </Card>
        </DynamicShadowWrapper>

        <div className="col-span-11 md:col-span-5 place-self-center w-full">
          <Carousel
            mediaItems={carouselMediaItems}
            onReorder={newOrder =>
              setOrderedMediaByTab(prev => ({ ...prev, [activeTab]: newOrder }))
            }
            selectedIds={selectedMedia[activeTab]}
            overriddenIds={Object.keys(mediaOverrides[activeTab] || {})}
            onSelectionChange={handleMediaSelectionChange}
            onMediaUpdate={handleMediaUpdate}
            onRevert={handleRevertMedia}
          />
        </div>
      </div>
    </div>
  );
}

