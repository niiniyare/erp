// Mermaid.js Configuration for AWO ERP Documentation
// Provides enhanced diagram rendering with custom styling and theme support

document.addEventListener('DOMContentLoaded', function() {
    'use strict';
    
    // Check if Mermaid is loaded
    if (typeof mermaid === 'undefined') {
        console.warn('Mermaid.js not loaded - diagrams may not render');
        return;
    }
    
    // Theme detection based on Material theme
    function getCurrentTheme() {
        const body = document.body;
        const isDark = body.getAttribute('data-md-color-scheme') === 'slate' ||
                      body.classList.contains('dark-theme') ||
                      window.matchMedia('(prefers-color-scheme: dark)').matches;
        return isDark ? 'dark' : 'default';
    }
    
    // Mermaid configuration
    const mermaidConfig = {
        startOnLoad: true,
        theme: getCurrentTheme(),
        securityLevel: 'loose', // Allow HTML in labels
        fontFamily: '"Roboto", "Helvetica Neue", Arial, sans-serif',
        fontSize: 14,
        
        // Flowchart configuration
        flowchart: {
            useMaxWidth: true,
            htmlLabels: true,
            curve: 'basis',
            nodeSpacing: 50,
            rankSpacing: 80,
            padding: 10
        },
        
        // Sequence diagram configuration
        sequence: {
            useMaxWidth: true,
            diagramMarginX: 50,
            diagramMarginY: 20,
            actorMargin: 50,
            width: 150,
            height: 65,
            boxMargin: 10,
            boxTextMargin: 5,
            noteMargin: 10,
            messageMargin: 35,
            messageAlign: 'center',
            mirrorActors: true,
            bottomMarginAdj: 1,
            wrap: false,
            wrapPadding: 10,
            labelBoxWidth: 50,
            labelBoxHeight: 20
        },
        
        // Class diagram configuration
        class: {
            useMaxWidth: true,
            defaultRenderer: 'dagre-wrapper',
            htmlLabels: true
        },
        
        // State diagram configuration
        state: {
            useMaxWidth: true,
            defaultRenderer: 'dagre-wrapper'
        },
        
        // Gantt chart configuration
        gantt: {
            useMaxWidth: true,
            leftPadding: 75,
            rightPadding: 20,
            gridLineStartPadding: 35,
            fontSize: 11,
            fontFamily: '"Roboto", "Helvetica Neue", Arial, sans-serif',
            sectionFontSize: 24,
            numberSectionStyles: 4
        },
        
        // Git graph configuration
        gitgraph: {
            useMaxWidth: true,
            mainBranchName: 'main',
            showBranches: true,
            showCommitLabel: true,
            rotateCommitLabel: false
        },
        
        // Entity Relationship diagram configuration
        er: {
            useMaxWidth: true,
            diagramPadding: 20,
            layoutDirection: 'TB',
            minEntityWidth: 100,
            minEntityHeight: 75,
            entityPadding: 15,
            stroke: 'gray',
            fill: 'honeydew',
            fontSize: 12
        },
        
        // User Journey configuration
        journey: {
            useMaxWidth: true,
            diagramMarginX: 50,
            diagramMarginY: 20,
            leftMargin: 150,
            width: 150,
            height: 50,
            boxMargin: 10,
            boxTextMargin: 5,
            noteMargin: 10,
            messageMargin: 35,
            messageAlign: 'center',
            bottomMarginAdj: 1,
            rightAngles: false,
            taskFontSize: 14,
            taskFontFamily: '"Roboto", "Helvetica Neue", Arial, sans-serif',
            taskMargin: 50,
            activationWidth: 10,
            textPlacement: 'fo',
            actorColours: ['#8085e9', '#ff6b6b', '#4ecdc4', '#45b7d1', '#96ceb4', '#ffeaa7', '#dda0dd', '#98d8c8'],
            sectionColours: ['#fff2cc', '#f8cecc', '#e1d5e7', '#dae8fc', '#d5e8d4', '#ffe6cc', '#fff2cc', '#f8cecc']
        },
        
        // Custom theme variables
        themeVariables: {
            // Primary colors
            primaryColor: '#3f51b5',
            primaryTextColor: '#ffffff',
            primaryBorderColor: '#3f51b5',
            
            // Background colors
            background: '#ffffff',
            secondaryColor: '#f5f5f5',
            tertiaryColor: '#fafafa',
            
            // Line colors
            lineColor: '#757575',
            
            // Text colors
            textColor: '#212121',
            
            // Node colors
            mainBkg: '#3f51b5',
            secondBkg: '#e8eaf6',
            tertiaryColor: '#c5cae9',
            
            // Specific diagram colors
            cScale0: '#3f51b5',
            cScale1: '#5c6bc0',
            cScale2: '#7986cb',
            cScale3: '#9fa8da',
            cScale4: '#c5cae9',
            cScale5: '#e8eaf6',
            
            // State colors
            specialStateColor: '#ff6b6b',
            
            // Class diagram colors
            classText: '#212121',
            
            // Sequence diagram colors
            actor0: '#3f51b5',
            actor1: '#5c6bc0',
            actor2: '#7986cb',
            actor3: '#9fa8da',
            
            // Git colors
            git0: '#4caf50',
            git1: '#2196f3',
            git2: '#ff9800',
            git3: '#f44336',
            git4: '#9c27b0',
            git5: '#607d8b',
            git6: '#795548',
            git7: '#009688',
            
            // Journey colors
            fillType0: '#3f51b5',
            fillType1: '#5c6bc0',
            fillType2: '#7986cb',
            fillType3: '#9fa8da',
            fillType4: '#c5cae9',
            fillType5: '#e8eaf6',
            fillType6: '#f3e5f5',
            fillType7: '#fce4ec'
        }
    };
    
    // Dark theme overrides
    const darkThemeConfig = {
        themeVariables: {
            background: '#263238',
            primaryColor: '#5c6bc0',
            primaryTextColor: '#ffffff',
            primaryBorderColor: '#5c6bc0',
            secondaryColor: '#37474f',
            tertiaryColor: '#455a64',
            lineColor: '#90a4ae',
            textColor: '#eceff1',
            mainBkg: '#5c6bc0',
            secondBkg: '#37474f',
            tertiaryColor: '#455a64',
            classText: '#eceff1'
        }
    };
    
    // Apply theme-specific configuration
    function updateMermaidTheme() {
        const currentTheme = getCurrentTheme();
        let config = { ...mermaidConfig };
        
        if (currentTheme === 'dark') {
            config.theme = 'dark';
            config.themeVariables = { ...config.themeVariables, ...darkThemeConfig.themeVariables };
        } else {
            config.theme = 'default';
        }
        
        mermaid.initialize(config);
    }
    
    // Initialize Mermaid
    updateMermaidTheme();
    
    // Watch for theme changes
    const observer = new MutationObserver(function(mutations) {
        mutations.forEach(function(mutation) {
            if (mutation.type === 'attributes' && 
                (mutation.attributeName === 'data-md-color-scheme' || 
                 mutation.attributeName === 'class')) {
                updateMermaidTheme();
                // Re-render all mermaid diagrams
                const diagrams = document.querySelectorAll('.mermaid');
                diagrams.forEach(function(diagram, index) {
                    const graphDefinition = diagram.textContent;
                    diagram.innerHTML = '';
                    diagram.removeAttribute('data-processed');
                    mermaid.render(`mermaidChart${index}`, graphDefinition).then(function(result) {
                        diagram.innerHTML = result.svg;
                    });
                });
            }
        });
    });
    
    // Start observing theme changes
    observer.observe(document.body, {
        attributes: true,
        attributeFilter: ['data-md-color-scheme', 'class']
    });
    
    // Enhanced diagram processing
    function enhanceDiagrams() {
        const diagrams = document.querySelectorAll('.mermaid');
        
        diagrams.forEach(function(diagram, index) {
            // Add wrapper for better styling
            if (!diagram.closest('.mermaid-wrapper')) {
                const wrapper = document.createElement('div');
                wrapper.className = 'mermaid-wrapper';
                wrapper.style.cssText = `
                    text-align: center;
                    margin: 20px 0;
                    padding: 15px;
                    border: 1px solid #e0e0e0;
                    border-radius: 8px;
                    background: #fafafa;
                    overflow-x: auto;
                `;
                
                diagram.parentNode.insertBefore(wrapper, diagram);
                wrapper.appendChild(diagram);
                
                // Add title if provided in comment
                const prevElement = wrapper.previousElementSibling;
                if (prevElement && prevElement.tagName === 'P' && 
                    prevElement.textContent.startsWith('<!-- title:')) {
                    const titleMatch = prevElement.textContent.match(/<!-- title:\s*(.+?)\s*-->/);
                    if (titleMatch) {
                        const title = document.createElement('h4');
                        title.textContent = titleMatch[1];
                        title.style.cssText = 'margin-bottom: 15px; color: #333;';
                        wrapper.insertBefore(title, diagram);
                        prevElement.style.display = 'none';
                    }
                }
            }
            
            // Add click-to-expand functionality for large diagrams
            diagram.addEventListener('click', function() {
                if (diagram.classList.contains('expanded')) {
                    diagram.classList.remove('expanded');
                    diagram.style.transform = '';
                    diagram.style.position = '';
                    diagram.style.zIndex = '';
                    diagram.style.background = '';
                    diagram.style.padding = '';
                    diagram.style.boxShadow = '';
                } else {
                    diagram.classList.add('expanded');
                    diagram.style.transform = 'scale(1.2)';
                    diagram.style.position = 'relative';
                    diagram.style.zIndex = '1000';
                    diagram.style.background = '#fff';
                    diagram.style.padding = '20px';
                    diagram.style.boxShadow = '0 10px 30px rgba(0,0,0,0.3)';
                }
            });
        });
    }
    
    // Process diagrams after they're rendered
    setTimeout(enhanceDiagrams, 1000);
    
    // Handle dynamic content loading
    const contentObserver = new MutationObserver(function(mutations) {
        mutations.forEach(function(mutation) {
            if (mutation.addedNodes.length > 0) {
                setTimeout(enhanceDiagrams, 500);
            }
        });
    });
    
    contentObserver.observe(document.body, {
        childList: true,
        subtree: true
    });
    
    // Add custom styles
    const mermaidStyles = document.createElement('style');
    mermaidStyles.textContent = `
        .mermaid-wrapper {
            position: relative;
            transition: all 0.3s ease;
        }
        
        .mermaid {
            cursor: pointer;
            transition: transform 0.3s ease;
        }
        
        .mermaid:hover {
            transform: scale(1.02);
        }
        
        .mermaid.expanded {
            cursor: zoom-out;
            border-radius: 8px;
        }
        
        /* Dark theme adjustments */
        [data-md-color-scheme="slate"] .mermaid-wrapper {
            background: #37474f;
            border-color: #546e7a;
        }
        
        [data-md-color-scheme="slate"] .mermaid.expanded {
            background: #263238 !important;
        }
        
        /* Mobile responsiveness */
        @media (max-width: 768px) {
            .mermaid-wrapper {
                margin: 15px -16px;
                padding: 10px;
                border-radius: 0;
                border-left: none;
                border-right: none;
            }
            
            .mermaid {
                max-width: 100%;
                overflow-x: auto;
            }
        }
        
        /* Print styles */
        @media print {
            .mermaid-wrapper {
                border: 1px solid #ccc;
                break-inside: avoid;
            }
            
            .mermaid {
                transform: none !important;
                position: static !important;
                z-index: auto !important;
                box-shadow: none !important;
            }
        }
    `;
    
    document.head.appendChild(mermaidStyles);
    
    // Expose configuration function for external use
    window.updateMermaidConfig = function(customConfig) {
        const mergedConfig = { ...mermaidConfig, ...customConfig };
        mermaid.initialize(mergedConfig);
    };
    
    // Debug logging
    console.log('Mermaid.js configured successfully for AWO ERP Documentation');
});