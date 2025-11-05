export type TabKey = 'twitter' | 'youtube' | 'instagram' | 'reddit' | 'mastodon' | 'artstation';

export type MediaItemType = { id: string; type: "image" | "video"; src: string; };

export type MediaOverride = { src: string; file: File; };

export type MediaOverrides = Record<TabKey, Record<string, MediaOverride>>;

export type FormDataState = {
  importData: {};
  twitter: { content: string };
  youtube: { title: string; description: string; tags: string };
  instagram: { caption: string };
  reddit: {  };
  mastodon: {  };
  artstation: {  };
};

// This type defines the props for our modular tab components
export type TabComponentProps<P extends keyof FormDataState> = {
  data: FormDataState[P];
  handleChange: <F extends keyof FormDataState[P]>(
    platform: P,
    field: F
  ) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
};