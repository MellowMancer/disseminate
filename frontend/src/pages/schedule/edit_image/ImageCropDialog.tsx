import React, { useState, useRef } from 'react';
import ReactCrop, { type Crop, type PixelCrop, convertToPixelCrop, centerCrop, makeAspectCrop } from 'react-image-crop';
import { Button } from '@/components/ui/button';
import { DialogClose } from '@radix-ui/react-dialog';
import { canvasPreview } from './canvasPreview';
import { useDebounceEffect } from './useDebounceEffect';
import 'react-image-crop/dist/ReactCrop.css';
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"

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

    const aspectOptions = [
        "Free",
        "(1 / 1)",
        "(3 / 2)",
        "(2 / 3)",
        "(16 / 9)",
        "(9 / 16)",
        "(4 / 3)",
        "(3 / 4)",
    ];

    const parseAspect = (input: string): number | undefined => {
        const match = input.match(/\((\d+)\s*\/\s*(\d+)\)/);
        if (!match) return undefined;
        const [, a, b] = match;
        setAspect(Number(a) / Number(b));
        if (imgRef.current) {
            const { width, height } = imgRef.current
            const newCrop = centerAspectCrop(width, height, Number (a) / Number (b))
            setCrop(newCrop)
            // Updates the preview
            setCompletedCrop(convertToPixelCrop(newCrop, width, height))
        }
        return;
    };
    // Center crop with aspect ratio
    function centerAspectCrop(mediaWidth: number, mediaHeight: number, aspect: number) {
        return centerCrop(
            makeAspectCrop({ unit: '%', width: 90 }, aspect, mediaWidth, mediaHeight),
            mediaWidth,
            mediaHeight
        );
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
        <div>
            <div className='grid gap-5 justify-center items-center grid-cols-1 md:grid-cols-2'>
                <div className="bg-opacity-50 flex justify-center items-center z-50">
                    <div className="max-w-lg w-full flex flex-col items-center">
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


                    </div>
                </div>
                <div>
                    <div>
                        <Label className="mb-2">Aspect Ratio</Label>
                        <RadioGroup
                            className="grid grid-cols-2 gap-2"
                            defaultValue="(1 / 1)"
                            onValueChange={(val: string) => parseAspect(val)}
                        >
                            {aspectOptions.map(opt => (
                                <div className="flex items-center space-x-2" key={opt}>
                                    <RadioGroupItem value={opt} id={opt} />
                                    <Label htmlFor={opt}>{opt}</Label>
                                </div>
                            ))}
                        </RadioGroup>
                    </div>
                </div>
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
    );
}
