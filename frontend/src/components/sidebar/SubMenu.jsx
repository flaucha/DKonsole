import React from 'react';

const SubMenu = ({ isOpen, children, animationStyle = 'slide', animationSpeed = 'medium' }) => {
    const getAnimationClasses = () => {
        switch (animationStyle) {
            case 'slide':
                return isOpen
                    ? 'max-h-[500px] opacity-100 translate-y-0'
                    : 'max-h-0 opacity-0 -translate-y-2';
            case 'fade':
                return isOpen
                    ? 'max-h-[500px] opacity-100'
                    : 'max-h-0 opacity-0';
            case 'scale':
                return isOpen
                    ? 'max-h-[500px] opacity-100 scale-100'
                    : 'max-h-0 opacity-0 scale-95';
            case 'rotate':
                return isOpen
                    ? 'max-h-[500px] opacity-100 rotate-0'
                    : 'max-h-0 opacity-0 rotate-[-2deg]';
            default:
                return isOpen
                    ? 'max-h-[500px] opacity-100 translate-y-0'
                    : 'max-h-0 opacity-0 -translate-y-2';
        }
    };

    const getSpeedDuration = () => {
        switch (animationSpeed) {
            case 'slow':
                return 'duration-500';
            case 'fast':
                return 'duration-150';
            case 'medium':
            default:
                return 'duration-300';
        }
    };

    const getTransitionClasses = () => {
        const speedClass = getSpeedDuration();
        switch (animationStyle) {
            case 'slide':
                return `transition-all ${speedClass} ease-out`;
            case 'fade':
                return `transition-all ${speedClass === 'duration-500' ? 'duration-400' : speedClass === 'duration-150' ? 'duration-200' : 'duration-250'} ease-in-out`;
            case 'scale':
                return `transition-all ${speedClass} ease-out transform-gpu`;
            case 'rotate':
                return `transition-all ${speedClass} ease-out transform-gpu origin-top-left`;
            default:
                return `transition-all ${speedClass} ease-out`;
        }
    };

    return (
        <div
            className={`overflow-hidden ${getTransitionClasses()} ${getAnimationClasses()}`}
            style={{
                transformOrigin: animationStyle === 'scale' ? 'top' : animationStyle === 'rotate' ? 'top left' : 'top'
            }}
        >
            <div className="space-y-1 mb-2">
                {children}
            </div>
        </div>
    );
};

export default SubMenu;
