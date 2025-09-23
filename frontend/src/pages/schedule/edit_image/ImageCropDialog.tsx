import React, { useState, useRef } from 'react';
import ReactCrop, { type Crop, type PixelCrop, convertToPixelCrop, centerCrop, makeAspectCrop } from 'react-image-crop';
import { Button } from '@/components/ui/button';
import { DialogClose } from '@radix-ui/react-dialog';
import { canvasPreview } from './canvasPreview';
import { useDebounceEffect } from './useDebounceEffect';
import 'react-image-crop/dist/ReactCrop.css'

type ImageCropDialogProps = {
    src: string;
    aspect?: number;
    onClose: () => void;
    onCropComplete: (croppedDataUrl: string) => void;
};

export function ImageCropDialog({ src, onClose, onCropComplete }: ImageCropDialogProps) {
    const previewCanvasRef = useRef<HTMLCanvasElement>(null)
    const imgRef = useRef<HTMLImageElement>(null)
    const [crop, setCrop] = useState<Crop>()
    const [completedCrop, setCompletedCrop] = useState<PixelCrop>()
    const [scale] = useState(1)
    const [rotate] = useState(0)
    const [aspect, setAspect] = useState<number | undefined>(1 / 1)


    // Center crop with aspect ratio
    function centerAspectCrop(mediaWidth: number, mediaHeight: number, aspect: number) {
        return centerCrop(
            makeAspectCrop({ unit: '%', width: 90 }, aspect, mediaWidth, mediaHeight),
            mediaWidth,
            mediaHeight
        );
    }

    function handleToggleAspectClick() {
        if (aspect) {
            setAspect(undefined)
        } else {
            setAspect(16 / 9)

            if (imgRef.current) {
                const { width, height } = imgRef.current
                const newCrop = centerAspectCrop(width, height, 16 / 9)
                setCrop(newCrop)
                // Updates the preview
                setCompletedCrop(convertToPixelCrop(newCrop, width, height))
            }
        }
    }

    // Update crop when image loads
    function onImageLoad(e: React.SyntheticEvent<HTMLImageElement>) {
        if (aspect) {
            const { width, height } = e.currentTarget;
            setCrop(centerAspectCrop(width, height, aspect));
        }
    }

    useDebounceEffect(
        async () => {
            if (
                completedCrop?.width &&
                completedCrop?.height &&
                imgRef.current &&
                previewCanvasRef.current
            ) {
                canvasPreview(
                    imgRef.current,
                    previewCanvasRef.current,
                    completedCrop,
                    scale,
                    rotate,
                )
            }
        },
        100,
        [completedCrop, scale, rotate],
    )

    // Export crop as base64 when user confirms
    function handleSave() {
        if (!previewCanvasRef.current) return;
        const dataUrl = previewCanvasRef.current.toDataURL('image/png');
        onCropComplete(dataUrl);
        onClose();
    }

    if (!open) return null;

    return (
        <div><div>
            <Button onClick={handleToggleAspectClick}>
                Toggle aspect {aspect ? 'off' : 'on'}
            </Button>
        </div>
            <div className="bg-opacity-50 flex justify-center items-center z-50">

                <div className="max-w-lg w-full">
                    {!!src && (
                        <ReactCrop
                            crop={crop}
                            onChange={(_, percentCrop) => setCrop(percentCrop)}
                            onComplete={(c) => setCompletedCrop(c)}
                            aspect={aspect}
                            minWidth={50}
                            minHeight={50}
                            ruleOfThirds
                            className='max-h-76 max-w-76'
                        >
                            <img
                                ref={imgRef}
                                alt="Crop me"
                                src={src}
                                style={{ transform: `scale(${scale}) rotate(${rotate}deg)` }}
                                onLoad={onImageLoad}
                                className='max-h-76 max-w-76'
                            />
                        </ReactCrop>
                    )}

                    <div className="mt-4">
                        {completedCrop && (
                            <canvas
                                ref={previewCanvasRef}
                                style={{
                                    border: '1px solid black',
                                    objectFit: 'contain',
                                    width: completedCrop.width,
                                    height: completedCrop.height,
                                    display: 'none'
                                }}
                            />
                        )}

                    </div>

                    <div className="mt-4 flex justify-end gap-2">
                        <DialogClose>
                            <Button onClick={onClose} >Cancel</Button>
                        </DialogClose>
                        <DialogClose>
                            <Button onClick={handleSave} variant="secondary">Save</Button>
                        </DialogClose>
                    </div>
                </div>
            </div>
        </div>
    );
}
