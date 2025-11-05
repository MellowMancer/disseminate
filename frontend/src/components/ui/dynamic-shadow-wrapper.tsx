import React, { isValidElement, useEffect, useRef, useState, type ReactNode } from 'react';
import * as NavigationMenuPrimitive from "@radix-ui/react-navigation-menu"

type DynamicShadowWrapperProps = {
    children: ReactNode;
    className?: string;
};

export function DynamicShadowWrapper({ children, className = '' }: Readonly<DynamicShadowWrapperProps>) {
    const ref = useRef<HTMLDivElement>(null);
    const [shadowOffset, setShadowOffset] = useState({ x: 0, y: 0 });
    const [windowWidth, setWindowWidth] = useState(globalThis.innerWidth);


    useEffect(() => {
        function handleResize() {
            setWindowWidth(globalThis.innerWidth);
        }
        globalThis.addEventListener('resize', handleResize);
        return () => globalThis.removeEventListener('resize', handleResize);
    }, []);

    let isNavMenuItem = false;

    if (React.Children.count(children) === 1) {
        const child = React.Children.only(children);
        if (isValidElement(child)) {
            isNavMenuItem = child.type === NavigationMenuPrimitive.Item;
        }
    }
    let scaleFactor: number;
    if (windowWidth >= 768) {
        scaleFactor = 1;
    } else if (windowWidth >= 640) {
        scaleFactor = 0.7;
    } else {
        scaleFactor = 0.5;
    }
    if (isNavMenuItem) {
        scaleFactor = 0.15;
    }

    useEffect(() => {
        function updateShadowPosition() {
            if (!ref.current) return;

            const rect = ref.current.getBoundingClientRect();

            let lightSourceX = globalThis.innerWidth / 2 - 120;
            let lightSourceY = 0;

            const elementCenterX = rect.left + rect.width / 2;
            const elementCenterY = rect.top + rect.height / 2;

            let offsetX = ((elementCenterX - lightSourceX) / lightSourceX) * 24;
            let offsetY = ((elementCenterY - lightSourceY) / lightSourceY) * 24;
            offsetX = Math.min(Math.max(offsetX, -24), 24)
            offsetY = Math.min(Math.max(offsetY, -24), 24)
            setShadowOffset({ x: offsetX * scaleFactor, y: offsetY * scaleFactor });
        }



        updateShadowPosition();
        globalThis.addEventListener('resize', updateShadowPosition);
        globalThis.addEventListener('scroll', updateShadowPosition, true);

        return () => {
            globalThis.removeEventListener('mousemove', updateShadowPosition);
            globalThis.removeEventListener('resize', updateShadowPosition);
            globalThis.removeEventListener('scroll', updateShadowPosition, true);
        };
    }, [windowWidth, scaleFactor]);

    return (
        <div
            ref={ref}
            className={`${className} transition-shadow duration-300 ease-in-out rounded-md`}
            style={{
                '--shadow': `${shadowOffset.x}px ${shadowOffset.y}px 0px 0px #1A1A1AA9`,
                boxShadow: 'var(--shadow)',
            } as React.CSSProperties}
        >
            {children}
        </div>
    );
}
