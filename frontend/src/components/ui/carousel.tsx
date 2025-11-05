import React, { useState, useEffect } from "react";
import { Button } from '@/components/ui/button';
import { GridIcon } from "lucide-react";
import {
    Dialog,
    DialogTrigger,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle
} from "@/components/ui/dialog";
import { ImageCropDialog } from "@/pages/schedule/edit_image/ImageCropDialog";

export type MediaItemType = { id: string; type: "image" | "video"; src: string; };

interface CarouselProps extends React.HTMLAttributes<HTMLDivElement> {
    mediaItems: MediaItemType[];
    onReorder: (newOrder: MediaItemType[]) => void;
    selectedIds: Set<string>;
    onSelectionChange: (id: string) => void;
    onMediaUpdate: (id: string, newSrc: string) => void;
    overriddenIds: string[];
    onRevert: (id: string) => void;
    fixedHeight?: number;
    maxWidthThreshold?: number;
}


interface SelectionButtonProps {
    itemId: string;
    isSelected: boolean;
    onSelectionChange: (id: string) => void;
}

export function SelectionButton({ itemId, isSelected, onSelectionChange }: Readonly<SelectionButtonProps>) {
    return (
        <button
            type="button"
            className={`absolute top-2 right-2 z-10 w-6 h-6 rounded-full border-2 border-white bg-black bg-opacity-50 flex items-center justify-center transition-all ${isSelected ? "bg-blue-500 border-blue-500" : "hover:bg-opacity-70"
                }`}
            onClick={() => onSelectionChange(itemId)}
            aria-pressed={isSelected}
            aria-label={isSelected ? "Deselect item" : "Select item"}
        >
            {isSelected && (
                <svg className="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                </svg>
            )}
        </button>
    );
}


// ---- Utility: Render media item preview ----
function MediaPreview({
    item,
    border,
    overrideLabel,
    playable = false
}: Readonly<{
    item: MediaItemType,
    border: string,
    overrideLabel?: boolean,
    playable?: boolean
}>) {
    return (
        <>
            {item.type === "image" ? (
                <img src={item.src} alt="" className={"h-full w-full object-contain mx-auto pointer-events-none" + border} draggable={false} />
            ) : (
                <video
                    src={item.src}
                    className={"h-full w-full object-contain mx-auto" + border}
                    controls={playable}
                    tabIndex={playable ? 0 : -1}
                    style={playable ? {} : { pointerEvents: "none" }}
                />
            )}
            {overrideLabel && (
                <div className="absolute top-8 left-2 bg-primary text-white text-xs font-bold px-2 py-1 rounded-full">
                    Edited
                </div>
            )}
        </>
    );
}

const transitionDuration = 180; // in ms

const Carousel: React.FC<CarouselProps> = ({
    mediaItems,
    onReorder,
    selectedIds,
    onSelectionChange,
    onMediaUpdate,
    fixedHeight = 400,
    maxWidthThreshold = 550,
    overriddenIds,
    onRevert,
    className,
}) => {
    const [currentIndex, setCurrentIndex] = useState(0);
    const [transitioning, setTransitioning] = useState(false);
    const [dimensions, setDimensions] = useState<(null | { width: number; height: number })[]>(mediaItems.map(() => null));
    const [containerHeight, setContainerHeight] = useState(fixedHeight);
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [orderedMedia, setOrderedMedia] = useState(mediaItems);

    useEffect(() => { setOrderedMedia(mediaItems); }, [mediaItems]);

    // ---- Carousel navigation ----
    const changeSlide = (newIndex: number) => {
        if (transitioning) return;
        setTransitioning(true);
        setTimeout(() => {
            setCurrentIndex(newIndex);
            setTransitioning(false);
        }, transitionDuration);
    };
    const prev = () => changeSlide(currentIndex === 0 ? orderedMedia.length - 1 : currentIndex - 1);
    const next = () => changeSlide(currentIndex === orderedMedia.length - 1 ? 0 : currentIndex + 1);

    // ---- Reorder logic ----
    const moveItem = (index: number, direction: "left" | "right") => {
        setOrderedMedia(prev => {
            const arr = [...prev];
            const targetIndex = direction === "left" ? index - 1 : index + 1;
            if (targetIndex < 0 || targetIndex >= arr.length) return arr;
            [arr[index], arr[targetIndex]] = [arr[targetIndex], arr[index]];
            onReorder(arr);
            return arr;
        });
    };

    useEffect(() => {
        setDimensions(mediaItems.map(() => null));
        let isCancelled = false;
        for (const [idx, item] of mediaItems.entries()) {
            const processDimensions = (width: number, height: number) => {
                if (!isCancelled) setDimensions(prev => { const n = [...prev]; n[idx] = { width, height }; return n; });
            };
            if (item.type === "image") {
                const img = new globalThis.Image();
                img.onload = () => processDimensions(img.naturalWidth, img.naturalHeight);
                img.src = item.src;
            } else {
                const video = document.createElement("video");
                video.onloadedmetadata = () => processDimensions(video.videoWidth, video.videoHeight);
                video.src = item.src;
            }
        }
        return () => { isCancelled = true; };
    }, [mediaItems]);

    useEffect(() => {
        const validDims = dimensions.filter((d): d is { width: number, height: number } => !!d);
        if (validDims.length !== mediaItems.length || !mediaItems.length) return;
        const widthsAtFixedHeight = validDims.map(d => (d.width / d.height) * fixedHeight);
        const maxWidth = Math.max(...widthsAtFixedHeight);
        setContainerHeight(maxWidth > maxWidthThreshold ? Math.floor(fixedHeight * (maxWidthThreshold / maxWidth)) : fixedHeight);
    }, [dimensions, fixedHeight, maxWidthThreshold, mediaItems.length]);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "ArrowLeft") prev();
            else if (e.key === "ArrowRight") next();
        };
        globalThis.addEventListener("keydown", handleKeyDown);
        return () => globalThis.removeEventListener("keydown", handleKeyDown);
    }, [currentIndex, orderedMedia.length]);

    if (!orderedMedia.length)
        return <div className="text-center text-gray-500">No media available</div>;

    const border = " bg-card border-border border-1 border-t-24 rounded-md shadow-(--shadow-override) md:shadow-(--shadow-override-md) lg:shadow-(--shadow-override-lg)";
    const current = orderedMedia[currentIndex];
    const isOverridden = overriddenIds.includes(current.id);
    const isSelected = selectedIds.has(current.id);

    // ---- Handle crop complete ----
    const handleCropComplete = (croppedSrc: string) => {
        if (editingIndex === null) return;
        onMediaUpdate(orderedMedia[editingIndex].id, croppedSrc);
        setEditingIndex(null);
    }

    // ---- Main Carousel and Dialog ----
    return (
        <div className={className + " duration-500 ease-in-out"}>
            <div className="mx-auto rounded-md relative overscroll-contain touch-none flex"
                style={{ maxWidth: maxWidthThreshold, height: containerHeight }}>
                <div
                    className="flex justify-center items-center h-full w-full relative group"
                    style={{ width: '100%' }}
                >
                    <MediaPreview item={current} border={border} overrideLabel={isOverridden} playable />
                    <SelectionButton itemId={current.id} isSelected={isSelected} onSelectionChange={onSelectionChange} />
                    {current.type === "image" &&
                        <div className="absolute bottom-3 left-1/2 -translate-x-1/2 flex gap-2 z-20 opacity-0 group-hover:opacity-100 transition-opacity">
                            <Dialog onOpenChange={open => !open && setEditingIndex(null)}>
                                <DialogTrigger asChild>
                                    <Button
                                        variant="secondary"
                                        size="sm"
                                        onClick={e => { e.stopPropagation(); setEditingIndex(currentIndex); }}>
                                        Edit
                                    </Button>
                                </DialogTrigger>
                                {editingIndex === currentIndex &&
                                    <DialogContent>
                                        <DialogHeader>
                                            <DialogTitle>Edit Image</DialogTitle>
                                            <DialogDescription>Crop, rotate, or adjust your image.</DialogDescription>
                                        </DialogHeader>
                                        <ImageCropDialog
                                            src={current.src}
                                            onClose={() => setEditingIndex(null)}
                                            onCropComplete={handleCropComplete}
                                        />
                                    </DialogContent>
                                }
                            </Dialog>
                            {isOverridden &&
                                <Button
                                    variant="destructive"
                                    size="sm"
                                    onClick={e => { e.stopPropagation(); onRevert(current.id); }}>
                                    Revert
                                </Button>
                            }
                        </div>
                    }
                </div>
            </div>
            <div className="relative mt-4 px-2 z-1 max-w-[550px] mx-auto flex justify-between items-center">
                <Button onClick={prev}>Prev</Button>
                <Dialog>
                    <DialogTrigger asChild>
                        <Button
                            className="flex items-center gap-2 text-primary-foreground bg-primary text-sm py-2 px-3 rounded-md cursor-pointer select-none hover:bg-primary/90 focus:ring-2 ring-primary ring-offset-2 transition-all"
                            aria-label="Show all media items"
                            tabIndex={0}
                            title="Click to view all media"
                        >
                            <GridIcon className="w-4 h-4 mr-1" aria-hidden="true" />
                            <span>
                                {currentIndex + 1} / {orderedMedia.length}
                            </span>
                        </Button>
                    </DialogTrigger>
                    <DialogContent className="max-w-4xl max-h-[80vh] overflow-auto">
                        <DialogHeader>
                            <DialogTitle>All Media Items</DialogTitle>
                            <DialogDescription>A preview of all your media files uploaded.</DialogDescription>
                        </DialogHeader>
                        <div className="grid grid-cols-4 gap-4 pt-2">
                            {orderedMedia.map((item, idx) => {
                                const isSelected = selectedIds.has(item.id);
                                return (
                                    <div key={item.id} className="relative border rounded flex flex-col items-center justify-center min-h-40 p-1">
                                        <div className="absolute top-1 right-1 flex gap-1">
                                            <button
                                                disabled={idx === 0}
                                                onClick={() => moveItem(idx, "left")}
                                                className="text-xs p-1 rounded bg-gray-100 disabled:opacity-50"
                                                aria-label="Move media left"
                                            >←</button>
                                            <button
                                                disabled={idx === orderedMedia.length - 1}
                                                onClick={() => moveItem(idx, "right")}
                                                className="text-xs p-1 rounded bg-gray-100 disabled:opacity-50"
                                                aria-label="Move media right"
                                            >→</button>
                                        </div>
                                        <button
                                            type="button"
                                            className={`border rounded p-1 flex flex-col items-center max-h-40 overflow-hidden focus:outline focus:outline-offset-2 focus:outline-primary ${isSelected ? "bg-primary text-primary-foreground font-semibold" : "border-border bg-transparent text-accent-foreground"}`}
                                            onClick={() => onSelectionChange(item.id)}
                                            aria-pressed={isSelected}
                                            aria-label={isSelected ? "Deselect media" : "Select media"}
                                        >
                                            <MediaPreview item={item} border="max-h-32" overrideLabel={overriddenIds.includes(item.id)} />
                                            <span className="mt-1 text-sm">{isSelected ? "Selected" : "Click to select"}</span>
                                        </button>
                                    </div>
                                );
                            })}
                        </div>
                    </DialogContent>
                </Dialog>
                <Button onClick={next}>Next</Button>
            </div>
        </div>
    );
};

export default Carousel;
