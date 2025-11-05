import React, { useState, useEffect } from "react";
import { Button } from '@/components/ui/button';
import { Dialog, DialogTrigger, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ImageCropDialog } from "@/pages/schedule/edit_image/ImageCropDialog";

export type MediaItemType = {
    id: string;
    type: "image" | "video";
    src: string;
};

type Dimensions = { width: number; height: number; };

interface CarouselProps extends React.HTMLAttributes<HTMLDivElement> {
    mediaItems: MediaItemType[];
    onReorder: (newOrder: MediaItemType[]) => void;
    selectedIds: Set<string>;
    onSelectionChange: (id: string) => void;
    onMediaUpdate: (id: string, newSrc: string) => void;
    fixedHeight?: number;
    maxWidthThreshold?: number;
    overriddenIds: string[];
    onRevert: (id: string) => void;
};

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
    const [dimensions, setDimensions] = useState<(Dimensions | null)[]>(mediaItems.map(() => null));
    const [containerHeight, setContainerHeight] = useState<number>(fixedHeight);
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [orderedMedia, setOrderedMedia] = React.useState(mediaItems);

    React.useEffect(() => {
        setOrderedMedia(mediaItems);
    }, [mediaItems]);

    const moveItem = (index: number, direction: "up" | "down") => {
        setOrderedMedia((prev) => {
            const newArray = [...prev];
            const targetIndex = direction === "up" ? index - 1 : index + 1;
            if (targetIndex < 0 || targetIndex >= newArray.length) return newArray;

            [newArray[index], newArray[targetIndex]] = [newArray[targetIndex], newArray[index]];

            // Notify parent about reorder
            onReorder(newArray);

            return newArray;
        });
    };

    const border = " bg-card border-border border-1 border-t-24 rounded-md shadow-(--shadow-override) md:shadow-(--shadow-override-md) lg:shadow-(--shadow-override-lg)";

    const changeSlide = (newIndex: number) => {
        if (transitioning) return;
        setTransitioning(true);
        setTimeout(() => {
            setCurrentIndex(newIndex);
            setTransitioning(false);
        }, transitionDuration);
    };

    const prev = () => {
        const newIndex = currentIndex === 0 ? mediaItems.length - 1 : currentIndex - 1;
        changeSlide(newIndex);
    };

    const next = () => {
        const newIndex = currentIndex === mediaItems.length - 1 ? 0 : currentIndex + 1;
        changeSlide(newIndex);
    };

    useEffect(() => {
        setDimensions(mediaItems.map(() => null));
        let isCancelled = false;

        for (const [idx, item] of mediaItems.entries()) {
            const processDimensions = (width: number, height: number) => {
                if (!isCancelled) {
                    setDimensions(prevDims => {
                        const newDims = [...prevDims];
                        newDims[idx] = { width, height };
                        return newDims;
                    });
                }
            };

            if (item.type === "image") {
                const img = new Image();
                img.onload = () => processDimensions(img.naturalWidth, img.naturalHeight);
                img.src = item.src;
            } else {
                const video = document.createElement("video");
                video.onloadedmetadata = () => processDimensions(video.videoWidth, video.videoHeight);
                video.src = item.src;
            }
        }

        return () => {
            isCancelled = true;
        };
    }, [mediaItems]);

    useEffect(() => {
        const validDims = dimensions.filter((d): d is Dimensions => d !== null);

        if (validDims.length !== mediaItems.length || mediaItems.length === 0) {
            return;
        }

        const widthsAtFixedHeight = validDims.map((d) => (d.width / d.height) * fixedHeight);
        const maxWidth = Math.max(...widthsAtFixedHeight);

        if (maxWidth > maxWidthThreshold) {
            const scaleFactor = maxWidthThreshold / maxWidth;
            const newHeight = Math.floor(fixedHeight * scaleFactor);
            setContainerHeight(newHeight);
        } else {
            setContainerHeight(fixedHeight);
        }
    }, [dimensions, fixedHeight, maxWidthThreshold, mediaItems.length]);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "ArrowLeft") prev();
            else if (e.key === "ArrowRight") next();
        };
        globalThis.addEventListener("keydown", handleKeyDown);
        return () => globalThis.removeEventListener("keydown", handleKeyDown);
    }, [currentIndex, mediaItems.length]);

    if (!mediaItems || mediaItems.length === 0) {
        return <div className="text-center text-gray-500">No media available</div>;
    }

    const handleCropComplete = (croppedSrc: string) => {
        if (editingIndex === null) return;
        const editedItem = mediaItems[editingIndex];
        onMediaUpdate(editedItem.id, croppedSrc);
        setEditingIndex(null);
    };

    return (
        <div className={className + " duration-500 ease-in-out"}>
            <div
                className="mx-auto rounded-md relative overscroll-contain touch-none flex"
                style={{ maxWidth: maxWidthThreshold, height: containerHeight }}
            >
                {mediaItems.map((item, idx) => {
                    const isSelected = selectedIds.has(item.id);
                    const isOverridden = overriddenIds.includes(item.id);
                    return (
                        <div
                            key={item.id}
                            className="flex justify-center items-center h-full w-full relative group"
                            style={{
                                transition: `opacity ${transitionDuration}ms ease`,
                                opacity: idx === currentIndex && !transitioning ? 1 : 0,
                                position: idx === currentIndex ? "relative" : "absolute",
                                pointerEvents: idx === currentIndex ? "auto" : "none",
                            }}
                        >
                            {item.type === "image" ? (
                                <img
                                    src={item.src}
                                    alt=""
                                    className={"h-full w-full object-contain mx-auto pointer-events-none" + border}
                                    draggable={false}
                                />
                            ) : (
                                <video
                                    src={item.src}
                                    className={"h-full w-full object-contain mx-auto" + border}
                                    controls
                                />
                            )}

                            {/* SELECTION CHECKBOX BUTTON */}
                            <button
                                type="button"
                                className={`absolute top-2 right-2 z-10 w-6 h-6 rounded-full border-2 border-white bg-black bg-opacity-50 flex items-center justify-center transition-all ${isSelected ? 'bg-blue-500 border-blue-500' : 'hover:bg-opacity-70'
                                    }`}
                                onClick={() => onSelectionChange(item.id)}
                                onKeyDown={(e) => {
                                    if (e.key === "Enter" || e.key === " ") {
                                        e.preventDefault();
                                        onSelectionChange(item.id);
                                    }
                                }}
                                aria-pressed={isSelected}
                                aria-label={isSelected ? "Deselect item" : "Select item"}
                            >
                                {isSelected && (
                                    <svg className="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                                    </svg>
                                )}
                            </button>

                            {isOverridden && (
                                <div className="absolute top-8 left-2 bg-primary text-white text-xs font-bold px-2 py-1 rounded-full">
                                    Edited
                                </div>
                            )}

                            {item.type === "image" && (
                                <div className="absolute bottom-3 left-1/2 -translate-x-1/2 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity z-20">
                                    <Dialog onOpenChange={(open) => !open && setEditingIndex(null)}>
                                        <DialogTrigger asChild>
                                            <Button
                                                variant="secondary"
                                                size="sm"
                                                onClick={(e) => { e.stopPropagation(); setEditingIndex(idx); }}
                                            >
                                                Edit
                                            </Button>
                                        </DialogTrigger>
                                        {editingIndex === idx && (
                                            <DialogContent>
                                                <DialogHeader>
                                                    <DialogTitle>Edit Image</DialogTitle>
                                                    <DialogDescription>Crop, rotate, or adjust your image.</DialogDescription>
                                                </DialogHeader>
                                                <ImageCropDialog
                                                    src={item.src}
                                                    onClose={() => setEditingIndex(null)}
                                                    onCropComplete={handleCropComplete}
                                                />
                                            </DialogContent>
                                        )}
                                    </Dialog>

                                    {isOverridden && (
                                        <Button
                                            variant="destructive"
                                            size="sm"
                                            onClick={(e) => { e.stopPropagation(); onRevert(item.id); }}
                                        >
                                            Revert
                                        </Button>
                                    )}
                                </div>
                            )}
                        </div>
                    );
                })}
            </div>

            <div className="relative mt-4 px-2 z-1 max-w-[550px] mx-auto">
                <div className="absolute left-0 space-x-2">
                    <Button onClick={prev}>Prev</Button>
                    <Button onClick={next}>Next</Button>
                </div>
                <Dialog>
                    <DialogTrigger asChild>
                        <div
                            className="text-primary-foreground bg-primary text-sm py-2 px-3 rounded-md absolute right-0 cursor-pointer select-none"
                            aria-label="Show all media items"
                        >
                            {currentIndex + 1} / {mediaItems.length}
                        </div>
                    </DialogTrigger>

                    <DialogContent className="max-w-4xl max-h-[80vh] overflow-auto">
                        <DialogHeader>
                            <DialogTitle>All Media Items</DialogTitle>
                            <DialogDescription>
                                A preview of all your media files uploaded.
                            </DialogDescription>
                        </DialogHeader>

                        <div className="grid grid-cols-4 gap-4 pt-2">
                            {orderedMedia.map((item, idx) => {
                                const isSelected = selectedIds.has(item.id);
                                return (
                                    <div key={item.id} className="relative border rounded flex flex-col items-center justify-center min-h-40 p-1">
                                        {/* Reorder buttons */}
                                        <div className="absolute top-1 right-1 flex gap-1">
                                            <button
                                                disabled={idx === 0}
                                                onClick={() => moveItem(idx, "up")}
                                                className="text-xs p-1 rounded bg-gray-100 disabled:opacity-50"
                                                aria-label="Move media up"
                                            >←</button>
                                            <button
                                                disabled={idx === orderedMedia.length - 1}
                                                onClick={() => moveItem(idx, "down")}
                                                className="text-xs p-1 rounded bg-gray-100 disabled:opacity-50"
                                                aria-label="Move media down"
                                            >→</button>
                                        </div>

                                        {/* Rendering the media and selection button etc */}
                                        <button
                                            type="button"
                                            className={`border rounded p-1 flex flex-col items-center max-h-40 overflow-hidden focus:outline focus:outline-offset-2 focus:outline-primary ${isSelected ? "bg-primary text-primary-foreground font-semibold" : "border-border bg-transparent text-accent-foreground"
                                                }`}
                                            onClick={() => onSelectionChange(item.id)}
                                            aria-pressed={isSelected}
                                            aria-label={isSelected ? "Deselect media" : "Select media"}
                                        >
                                            {item.type === "image" ? (
                                                <img src={item.src} alt="" className="max-h-32 object-contain select-none" draggable={false} />
                                            ) : (
                                                <video src={item.src} className="max-h-32 object-contain select-none" />
                                            )}
                                            <span className="mt-1 text-sm">{isSelected ? "Selected" : "Click to select"}</span>
                                        </button>
                                    </div>
                                );
                            })}
                        </div>
                    </DialogContent>
                </Dialog>
            </div >
        </div >
    );
};

export default Carousel;
