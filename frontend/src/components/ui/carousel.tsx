import React, { useState, useEffect } from "react";
import { Button } from '@/components/ui/button';
import { DialogTrigger, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import  { ImageCropDialog }  from "@/pages/schedule/edit_image/ImageCropDialog";

type MediaItemType = {
    id: string,
    type: "image" | "video";
    src: string;
  };

type Dimensions = {
    width: number;
    height: number;
};

interface CarouselProps extends React.HTMLAttributes<HTMLDivElement> {
    mediaItems: MediaItemType[];
    fixedHeight?: number;   // e.g., 400px
    maxWidthThreshold?: number; // e.g., 700px max allowed width
};

const transitionDuration = 180; // in ms

const Carousel: React.FC<CarouselProps> = ({
    mediaItems,
    fixedHeight = 400,
    maxWidthThreshold = 550,
    className,
}) => {
    const [currentIndex, setCurrentIndex] = useState(0);
    const [transitioning, setTransitioning] = useState(false);
    const [dimensions, setDimensions] = useState<(Dimensions | null)[]>(mediaItems.map(() => null));
    const [containerHeight, setContainerHeight] = useState<number>(fixedHeight);
    const border = " bg-card border-card-outline border-1 border-t-24 rounded-md shadow-(--shadow-override) md:shadow-(--shadow-override-md) lg:shadow-(--shadow-override-lg)";
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [mediaList, setMediaList] = useState<MediaItemType[]>(mediaItems);

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
        mediaItems.forEach((item, idx) => {
            if (item.type === "image") {
                const img = new Image();
                img.onload = () => {
                    setDimensions((dims) => {
                        const newDims = [...dims];
                        newDims[idx] = { width: img.naturalWidth, height: img.naturalHeight };
                        return newDims;
                    });
                };
                img.src = item.src;
            } else {
                const video = document.createElement("video");
                video.onloadedmetadata = () => {
                    setDimensions((dims) => {
                        const newDims = [...dims];
                        newDims[idx] = { width: video.videoWidth, height: video.videoHeight };
                        return newDims;
                    });
                };
                video.src = item.src;
            }
        });
    }, [mediaItems]);

    // Calculate scaling and adjust container height if needed
    useEffect(() => {
        if (dimensions.some((d) => d === null)) return;

        const dims = dimensions as Dimensions[];

        // Calculate widths at fixedHeight
        const widthsAtFixedHeight = dims.map((d) => (d.width / d.height) * fixedHeight);

        const maxWidth = Math.max(...widthsAtFixedHeight);

        if (maxWidth > maxWidthThreshold) {
            // Scale down height to keep max width = maxWidthThreshold
            const scaleFactor = maxWidthThreshold / maxWidth;
            const newHeight = Math.floor(fixedHeight * scaleFactor);
            setContainerHeight(newHeight);
        } else {
            setContainerHeight(fixedHeight);
        }
    }, [dimensions, fixedHeight, maxWidthThreshold]);

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
        const updatedMedia = [...mediaList];
        updatedMedia[editingIndex] = { ...updatedMedia[editingIndex], src: croppedSrc };
        setMediaList(updatedMedia);
        setEditingIndex(null); // close dialog
    };


    return (
        <div className={className + "duration-500 ease-in-out"}>
            <div
                className="mx-auto rounded-md relative overscroll-contain touch-none flex"
                style={{ maxWidth: maxWidthThreshold, height: containerHeight }}
            >
                {mediaList.map((item, idx) => (
                    <div
                        key={idx}
                        className="flex justify-center items-center h-full w-full"
                        style={{
                            transition: `opacity ${transitionDuration}ms ease`,
                            opacity: idx === currentIndex && !transitioning ? 1 : 0,
                            position: idx === currentIndex ? "relative" : "absolute",
                            pointerEvents: idx === currentIndex ? "auto" : "none",
                        }}
                    >
                        <DialogTrigger className="w-full h-full">
                            
                            {item.type === "image" ? (
                                <img
                                    src={item.src}
                                    alt=""
                                    className={"h-full w-auto object-contain mx-auto" + border}
                                    onClick={() => setEditingIndex(idx)}
                                    style={{ cursor: "pointer" }}
                                />
                            ) : (
                                <video
                                    src={item.src}

                                    className={"h-full w-auto object-contain mx-auto" + border}
                                />
                            )}
                            
                        </DialogTrigger>
                        {editingIndex === idx && item.type === "image" && (
                            <DialogContent>
                                <DialogHeader>
                                    <DialogTitle>Edit Image</DialogTitle>
                                    <DialogDescription>
                                        {/* Place your cropping component here */}
                                    </DialogDescription>
                                </DialogHeader>
                                <ImageCropDialog
                                    src={mediaList[idx].src}
                                    onClose={() => setEditingIndex(null)}
                                    onCropComplete={handleCropComplete}
                                />
                            </DialogContent>
                        )}
                    </div>
                ))}
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
