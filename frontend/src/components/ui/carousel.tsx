import React, { useState, useEffect } from "react";
import { Button } from '@/components/ui/button';
import { Dialog, DialogTrigger, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ImageCropDialog } from "@/pages/schedule/edit_image/ImageCropDialog";

// Keep this type definition consistent
export type MediaItemType = {
    id: string;
    type: "image" | "video";
    src: string;
};

type Dimensions = { width: number; height: number; };

// --- ADD NEW PROPS to the interface ---
interface CarouselProps extends React.HTMLAttributes<HTMLDivElement> {
    mediaItems: MediaItemType[];
    selectedIds: Set<string>;                  // <-- NEW: Receives the set of selected IDs
    onSelectionChange: (id: string) => void;   // <-- NEW: Callback for selection changes
    onMediaUpdate: (id: string, newSrc: string) => void; // <-- NEW: Callback for when an image is cropped
    fixedHeight?: number;
    maxWidthThreshold?: number;
    overriddenIds: string[]; // <-- NEW: An array of IDs that have been edited for this tab
    onRevert: (id: string) => void;
};

const transitionDuration = 180; // in ms

const Carousel: React.FC<CarouselProps> = ({
    mediaItems,
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
    const border = " bg-card border-card-outline border-1 border-t-24 rounded-md shadow-(--shadow-override) md:shadow-(--shadow-override-md) lg:shadow-(--shadow-override-lg)";

    // Transition wrapping for slide changes
    const changeSlide = (newIndex: number) => {
        if (transitioning) return; // Avoid triggering multiple times during transition
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

    // Preload media intrinsic sizes
    useEffect(() => {
        setDimensions(mediaItems.map(() => null));
        let isCancelled = false;

        for (const [idx, item] of mediaItems.entries()) {
            const processDimensions = (width: number, height: number) => {
                if (!isCancelled) {
                    setDimensions(prevDims => {
                        // Use a callback to prevent stale state issues
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
        };

        return () => {
            isCancelled = true; // Cleanup function
        };
    }, [mediaItems]);

    // Calculate scaling and adjust container height if needed
    useEffect(() => {
        const validDims = dimensions.filter((d): d is Dimensions => d !== null);

        // If not all dimensions have been loaded yet, do nothing.
        if (validDims.length !== mediaItems.length || mediaItems.length === 0) {
            return;
        }

        // Now, we can safely work with validDims
        const widthsAtFixedHeight = validDims.map((d) => (d.width / d.height) * fixedHeight);
        const maxWidth = Math.max(...widthsAtFixedHeight);

        if (maxWidth > maxWidthThreshold) {
            const scaleFactor = maxWidthThreshold / maxWidth;
            const newHeight = Math.floor(fixedHeight * scaleFactor);
            setContainerHeight(newHeight);
        } else {
            setContainerHeight(fixedHeight);
        }
        // This hook now correctly depends on all its inputs
    }, [dimensions, fixedHeight, maxWidthThreshold, mediaItems.length]);

    // Keyboard arrow key navigation with transition
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "ArrowLeft") prev();
            else if (e.key === "ArrowRight") next();
        };
        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [currentIndex, mediaItems.length]);


    if (!mediaItems || mediaItems.length === 0) {
        return <div className="text-center text-gray-500">No media available</div>;
    }

    const handleCropComplete = (croppedSrc: string) => {
        if (editingIndex === null) return;
        const editedItem = mediaItems[editingIndex];
        onMediaUpdate(editedItem.id, croppedSrc); // Lift the state up to the parent
        setEditingIndex(null); // close dialog
    };


    return (
        <div className={className + " duration-500 ease-in-out"}>
            <div
                className="mx-auto rounded-md relative overscroll-contain touch-none flex"
                style={{ maxWidth: maxWidthThreshold, height: containerHeight }}
            >
                {mediaItems.map((item, idx) => {
                    const isSelected = selectedIds.has(item.id);
                    const isOverridden = overriddenIds.includes(item.id); // Check if the item is selected
                    return (
                        <div
                            key={item.id}
                            className="flex justify-center items-center h-full w-full"
                            style={{
                                transition: `opacity ${transitionDuration}ms ease`,
                                opacity: idx === currentIndex && !transitioning ? 1 : 0,
                                position: idx === currentIndex ? "relative" : "absolute",
                                pointerEvents: idx === currentIndex ? "auto" : "none",
                            }}
                        >
                            {/* WRAPPER for selection click and overlay */}
                            <button
                                type="button"
                                className="relative w-full h-full cursor-pointer group"
                                onClick={() => onSelectionChange(item.id)} // Make the whole item selectable
                                onKeyDown={(e) => {
                                    if (e.key === "Enter" || e.key === " ") {
                                        e.preventDefault();
                                        onSelectionChange(item.id);
                                    }
                                }}
                                aria-pressed={isSelected}
                                tabIndex={0}
                                style={{ background: "none", border: "none", padding: 0 }}
                            >
                                {isOverridden && (
                                    <div className="absolute top-8 left-2 bg- text-white text-xs font-bold px-2 py-1 rounded-full">
                                        Edited
                                    </div>
                                )}
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
                                        className={"h-full w-full object-contain mx-auto pointer-events-none" + border}
                                    />
                                )}

                                {/* SELECTION CHECKBOX OVERLAY */}
                                <div
                                    className={`absolute top-7 right-3 w-6 h-6 rounded-full border-2 border-white bg-black bg-opacity-50 flex items-center justify-center transition-all ${isSelected ? 'bg-blue-500 border-blue-500' : 'group-hover:bg-opacity-70'
                                        }`}
                                >
                                    {isSelected && (
                                        <svg className="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                                        </svg>
                                    )}
                                </div>

                                {item.type === "image" && (
                                    <div className="absolute bottom-3 left-1/2 -translate-x-1/2 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
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
                            </button>
                        </div>
                    );
                })}
            </div>

            <div className="relative mt-4 px-2 z-1 max-w-[550px] mx-auto">
                <div className="absolute left-0 space-x-2">
                    <Button onClick={prev}>Prev</Button>
                    <Button onClick={next}>Next</Button>
                </div>
                <div className="text-white bg-highlight text-sm py-2 px-3 rounded-md absolute right-0">
                    {currentIndex + 1} / {mediaItems.length}
                </div>
            </div>


        </div >
    );
};

export default Carousel;
