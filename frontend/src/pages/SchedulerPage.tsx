import React, { useState, useEffect, useMemo } from 'react';
import axios from 'axios';
import { toast } from 'sonner';
import { useLocation, useNavigate } from 'react-router-dom';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { DynamicShadowWrapper } from "@/components/ui/dynamic-shadow-wrapper";
import Carousel from "@/components/ui/carousel";

import type { FormDataState } from '@/types/forms';
import { TwitterTab } from '@/pages/schedule/TwitterTab';
import { YouTubeTab } from '@/pages/schedule/YoutubeTab';
import { InstagramTab } from './schedule/InstagramTab';

type TabKey = 'twitter' | 'youtube' | 'instagram' | 'reddit' | 'mastodon' | 'artstation';

type MediaItemType = {
  id: string;
  type: "image" | "video";
  src: string;
};

const initialFormData: FormDataState = {
  importData: {},
  twitter: { content: '' },
  youtube: { title: '', description: '', tags: '' },
  instagram: { caption: '' },
  reddit: {},
  mastodon: {},
  artstation: {},
};

const createInitialSelectedMedia = (): Record<TabKey, Set<string>> => ({
  twitter: new Set(),
  youtube: new Set(),
  instagram: new Set(),
  reddit: new Set(),
  mastodon: new Set(),
  artstation: new Set(),
});

type MediaOverride = { src: string; file: File; };
type MediaOverrides = Record<TabKey, Record<string, MediaOverride>>;
const createInitialMediaOverrides = (): MediaOverrides => ({
  twitter: {}, youtube: {}, instagram: {}, reddit: {}, mastodon: {}, artstation: {},
});


export function SchedulerPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const [isReady, setIsReady] = useState(false);
  const [formData, setFormData] = useState<FormDataState>(initialFormData);
  const [activeTab, setActiveTab] = useState<TabKey>('twitter');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [originalMediaItems, setOriginalMediaItems] = useState<MediaItemType[]>([]);
  const [originalFileMap, setOriginalFileMap] = useState<Map<string, File>>(new Map());
  const [mediaOverrides, setMediaOverrides] = useState<MediaOverrides>(createInitialMediaOverrides());
  const [selectedMedia, setSelectedMedia] = useState(createInitialSelectedMedia());


  useEffect(() => {
    const files = location.state?.files as FileList | undefined;

    if (!files || files.length === 0) {
      console.error("SchedulerPage loaded without files.");
      return;
    }

    const items: MediaItemType[] = [];
    const map = new Map<string, File>();
    for (const file of Array.from(files)) {
      const id = `${file.name}-${file.lastModified}-${file.size}`;
      items.push({
        id,
        type: file.type.startsWith("video") ? "video" : "image",
        src: URL.createObjectURL(file),
      });
      map.set(id, file);
    };

    setOriginalMediaItems(items);
    setOriginalFileMap(map);
    setIsReady(true);

    return () => {
      for (const item of items) { URL.revokeObjectURL(item.src) }
    };
  }, [location.state?.files]);

  const handleChange = <P extends keyof FormDataState>(
    platform: P,
    field: keyof FormDataState[P]
  ) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData(prev => ({
      ...prev,
      [platform]: { ...prev[platform], [field]: event.target.value },
    }));
  };

  const carouselMediaItems = useMemo(() => {
    const overridesForTab = mediaOverrides[activeTab] || {};
    return originalMediaItems.map(item => {
      if (overridesForTab[item.id]) {
        // If an override exists for this tab, use its src
        return { ...item, src: overridesForTab[item.id].src };
      }
      // Otherwise, use the original
      return item;
    });
  }, [originalMediaItems, mediaOverrides, activeTab]);


  const handleMediaSelectionChange = (mediaId: string) => {
    setSelectedMedia(prev => {
      const newSelectionsForTab = new Set(prev[activeTab]); // Copy the set for the active tab
      if (newSelectionsForTab.has(mediaId)) {
        newSelectionsForTab.delete(mediaId);
      } else {
        newSelectionsForTab.add(mediaId);
      }
      return { ...prev, [activeTab]: newSelectionsForTab };
    });
  };

  const handleMediaUpdate = async (id: string, newSrc: string) => {
    const response = await fetch(newSrc);
    const blob = await response.blob();
    const originalFile = originalFileMap.get(id);
    const newFile = new File([blob], originalFile?.name || 'cropped-image.png', { type: blob.type });

    const newOverride: MediaOverride = { src: newSrc, file: newFile };

    // Update the overrides state for the currently active tab
    setMediaOverrides(prev => ({
      ...prev,
      [activeTab]: {
        ...prev[activeTab],
        [id]: newOverride,
      },
    }));
    toast.success(`Edit saved for ${activeTab}.`);
  };

  const handleRevertMedia = (id: string) => {
    setMediaOverrides(prev => {
      const newOverridesForTab = { ...prev[activeTab] };
      delete newOverridesForTab[id]; // Remove the override for this item
      return {
        ...prev,
        [activeTab]: newOverridesForTab,
      };
    });
    toast.info("Reverted to original image.");
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setIsSubmitting(true);

    const selectedIds = selectedMedia[activeTab];
    if (selectedIds.size === 0) {
      toast.error(`Please select at least one media item for ${activeTab}.`);
      setIsSubmitting(false);
      return;
    }


    const submissionData = new FormData();
    const platformData = formData[activeTab];
    submissionData.append('platform', activeTab);
    submissionData.append('platformData', JSON.stringify(platformData));

    // 5. Append only the selected files
    const overridesForTab = mediaOverrides[activeTab] || {};

    for (const id of selectedIds) {
      let fileToSubmit: File | undefined;

      // Check if there is an override for this platform
      if (overridesForTab[id]) {
        fileToSubmit = overridesForTab[id].file;
      } else {
        // Otherwise, fall back to the original file
        fileToSubmit = originalFileMap.get(id);
      }

      if (fileToSubmit) {
        submissionData.append('media', fileToSubmit);
      }
    };

    try {
      // The actual API call to your backend
      await axios.post('/api/create', submissionData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      toast.success('Posted successfully!');
    } catch (error) {
      toast.error('Failed to create post.');
      console.error('Submission Error:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  useEffect(() => {
    if (originalMediaItems.length === 0) return;

    setSelectedMedia(prev => {
      // Copy previous to keep selections for other tabs
      const newSelected = { ...prev };
      // Select all media IDs on active tab
      newSelected[activeTab] = new Set(originalMediaItems.map(item => item.id));
      return newSelected;
    });
  }, [originalMediaItems, activeTab]);

  if (!isReady) {
    return (
      <div className='w-full h-full grid place-items-center p-4'>
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Loading Media...</CardTitle>
            <CardDescription>
              If you refreshed this page, the media files were lost and you'll need to go back.
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

  return (
    <div className='w-full h-full grid place-items-center p-4'>
      <div className="grid grid-cols-11 grid-rows-min gap-y-8 md:gap-x-8 w-full h-full content-center max-w-6xl">
        <DynamicShadowWrapper className="col-span-11 md:col-span-6 h-min ">
          <Card>
            <CardHeader>
              <CardTitle>Create a New Post</CardTitle>
              <CardDescription>Fill out the details for each platform you want to post to.</CardDescription>
            </CardHeader>

            <CardContent>
              <form onSubmit={handleSubmit}>
                <Tabs
                  value={activeTab}
                  onValueChange={(value) => setActiveTab(value as TabKey)}
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
                  <Button type="submit" className="w-full bg-primary text-primary-foreground" disabled={isSubmitting}>
                    {isSubmitting ? 'Posting...' : `Post for ${activeTab}`}
                  </Button>
                </CardFooter>
              </form>
            </CardContent>
          </Card>
        </DynamicShadowWrapper>
        <div className='col-span-11 md:col-span-5 place-self-center w-full'>
          <Carousel
            mediaItems={carouselMediaItems}
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