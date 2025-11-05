import { useState, useEffect, useMemo } from 'react';
import axios from 'axios';
import { toast } from 'sonner';
import type { FormDataState, TabKey, MediaItemType, MediaOverride } from '@/types/types';

export function useMediaManager(files: FileList | undefined, activeTab: TabKey) {
  const [isReady, setIsReady] = useState(false);
  const [formData, setFormData] = useState<FormDataState>({
    importData: {},
    twitter: { content: '' },
    youtube: { title: '', description: '', tags: '' },
    instagram: { caption: '' },
    reddit: {},
    mastodon: {},
    artstation: {},
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [originalMediaItems, setOriginalMediaItems] = useState<MediaItemType[]>([]);
  const [originalFileMap, setOriginalFileMap] = useState<Map<string, File>>(new Map());
  const [mediaOverrides, setMediaOverrides] = useState<Record<TabKey, Record<string, MediaOverride>>>({
    twitter: {},
    youtube: {},
    instagram: {},
    reddit: {},
    mastodon: {},
    artstation: {},
  });
  const [selectedMedia, setSelectedMedia] = useState<Record<TabKey, Set<string>>>({
    twitter: new Set(),
    youtube: new Set(),
    instagram: new Set(),
    reddit: new Set(),
    mastodon: new Set(),
    artstation: new Set(),
  });
  const [orderedMediaByTab, setOrderedMediaByTab] = useState<Record<TabKey, MediaItemType[]>>({
    twitter: [],
    youtube: [],
    instagram: [],
    reddit: [],
    mastodon: [],
    artstation: [],
  });

  useEffect(() => {
    if (!files || files.length === 0) return;

    const items: MediaItemType[] = [];
    const map = new Map<string, File>();

    for (const file of Array.from(files)) {
      const id = `${file.name}-${file.lastModified}-${file.size}`;
      items.push({
        id,
        type: file.type.startsWith('video') ? 'video' : 'image',
        src: URL.createObjectURL(file),
      });
      map.set(id, file);
    }

    setOriginalMediaItems(items);
    setOriginalFileMap(map);
    setIsReady(true);

    return () => {
      for (const item of items) URL.revokeObjectURL(item.src);
    };
  }, [files]);

  useEffect(() => {
    if (originalMediaItems.length === 0) return;
    setOrderedMediaByTab(prev => {
      const newOrder = { ...prev };
      (Object.keys(newOrder) as TabKey[]).forEach(tab => {
        if (!newOrder[tab] || newOrder[tab].length === 0) newOrder[tab] = originalMediaItems;
      });
      return newOrder;
    });

    setSelectedMedia(prev => {
      const newSelected = { ...prev };
      (Object.keys(newSelected) as TabKey[]).forEach(tab => {
        if (!newSelected[tab] || newSelected[tab].size === 0) {
          newSelected[tab] = new Set(originalMediaItems.map(item => item.id));
        }
      });
      return newSelected;
    });
  }, [originalMediaItems]);

  const carouselMediaItems = useMemo(() => {
    const ordered = orderedMediaByTab?.[activeTab] || [];
    const overridesForTab = mediaOverrides?.[activeTab] || {};
    return ordered.map(item => (overridesForTab[item.id] ? { ...item, src: overridesForTab[item.id].src } : item));
  }, [orderedMediaByTab, mediaOverrides, activeTab]);

  // Handlers - selection, update, revert
  const handleMediaSelectionChange = (mediaId: string) => {
    setSelectedMedia(prev => {
      const newSet = new Set(prev[activeTab]);
      if (newSet.has(mediaId)) newSet.delete(mediaId);
      else newSet.add(mediaId);
      return { ...prev, [activeTab]: newSet };
    });
  };

  const handleMediaUpdate = async (id: string, newSrc: string) => {
    const response = await fetch(newSrc);
    const blob = await response.blob();
    const originalFile = originalFileMap.get(id);
    const newFile = new File([blob], originalFile?.name || 'cropped-image.png', { type: blob.type });

    const newOverride: MediaOverride = { src: newSrc, file: newFile };

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
      delete newOverridesForTab[id];
      return {
        ...prev,
        [activeTab]: newOverridesForTab,
      };
    });
    toast.info('Reverted to original image.');
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

    const overridesForTab = mediaOverrides[activeTab] || {};

    for (const id of orderedMediaByTab[activeTab] || []) {
      if (!selectedIds.has(id.id)) continue;
      const override = overridesForTab[id.id];
      const fileToSubmit = override ? override.file : originalFileMap.get(id.id);
      if (fileToSubmit) submissionData.append('media', fileToSubmit);
    }

    try {
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

  return {
    isReady,
    formData,
    setFormData,
    isSubmitting,
    orderedMediaByTab,
    setOrderedMediaByTab,
    mediaOverrides,
    selectedMedia,
    handleMediaSelectionChange,
    handleMediaUpdate,
    handleRevertMedia,
    handleSubmit,
    carouselMediaItems,
  };
}
