import React, { useState, useEffect, useRef } from "react";
import { Button } from '@/components/ui/button';

type MediaItemType = {
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

    return (
        <div className={className + "duration-500 ease-in-out"}>
            <div
                className="mx-auto rounded-md relative overscroll-contain touch-none flex"
                style={{ maxWidth: maxWidthThreshold, height: containerHeight }}
            >
                {mediaItems.map((item, idx) => (
                    <div
                        key={idx}
                        // Use opacity and pointer events for fade transition
                        style={{
                            transition: `opacity ${transitionDuration}ms ease`,
                            opacity: idx === currentIndex && !transitioning ? 1 : 0,
                            position: idx === currentIndex ? "relative" : "absolute",
                            width: "100%",
                            height: "100%",
                            display: "flex",
                            justifyContent: "center",
                            alignItems: "center",
                            pointerEvents: idx === currentIndex ? "auto" : "none",
                        }}
                    >
                        {item.type === "image" ? (
                            <img
                                src={item.src}
                                alt=""
                                className={"h-full w-auto object-contain mx-auto" + border}
                            />
                        ) : (
                            <video
                                src={item.src}
                                controls
                                className={"h-full w-auto object-contain mx-auto" + border}
                            />
                        )}
                    </div>
                ))}
            </div>

            <div className="relative mt-4 px-2 z-100 max-w-[550px] mx-auto">
                <div className="absolute left-0 space-x-2">
                    <Button onClick={prev}>Prev</Button>
                    <Button onClick={next}>Next</Button>
                </div>
                <div className="text-white bg-highlight text-sm py-2 px-3 rounded-md absolute right-0">
                    {currentIndex + 1} / {mediaItems.length}
                </div>
            </div>


        </div>
    );
};

export default Carousel;
