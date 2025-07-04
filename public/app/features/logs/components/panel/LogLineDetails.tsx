import { css } from '@emotion/css';
import { Resizable } from 're-resizable';
import { useCallback, useEffect, useMemo, useRef } from 'react';

import { GrafanaTheme2 } from '@grafana/data';
import { getDragStyles, useStyles2 } from '@grafana/ui';

import { LogLineDetailsComponent } from './LogLineDetailsComponent';
import { useLogListContext } from './LogListContext';
import { LogListModel } from './processing';
import { LOG_LIST_MIN_HEIGHT, LOG_LIST_MIN_WIDTH } from './virtualization';

export interface Props {
  containerElement: HTMLDivElement;
  focusLogLine: (log: LogListModel) => void;
  logs: LogListModel[];
  onResize(): void;
}

export type LogLineDetailsPosition = 'bottom' | 'right';

export const DETAILS_BOTTOM_HEIGHT = 40;

export const LogLineDetails = ({ containerElement, focusLogLine, logs, onResize }: Props) => {
  const { detailsHeight, detailsPosition, detailsWidth, logOptionsStorageKey, setDetailsHeight, setDetailsWidth, showDetails } =
    useLogListContext();
  const styles = useStyles2(getStyles);
  const dragStyles = useStyles2(getDragStyles);
  const containerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    focusLogLine(showDetails[0]);
    // Just once
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleResize = useCallback(() => {
    if (!containerRef.current) {
      return;
    }
    if (detailsPosition === 'right') {
      setDetailsWidth(containerRef.current.clientWidth);
    } else {
      setDetailsHeight(containerRef.current.clientHeight);
    }
    onResize();
  }, [detailsPosition, onResize, setDetailsHeight, setDetailsWidth]);

  const resizable = useMemo(() => {
    if (detailsPosition === 'right') {
      return {
        defaultSize: { width: detailsWidth, height: containerElement.clientHeight },
        enable: { left: true },
        handleClasses: { left: dragStyles.dragHandleVertical },
        maxHeight: undefined,
        maxWidth: containerElement.clientWidth - LOG_LIST_MIN_WIDTH,
        size: { width: detailsWidth, height: containerElement.clientHeight },
      };
    }

    return {
      defaultSize: { width: '100%', height: detailsHeight },
      enable: { top: true },
      handleClasses: { top: dragStyles.dragHandleHorizontal },
      maxHeight: containerElement.clientHeight - LOG_LIST_MIN_HEIGHT,
      maxWidth: undefined,
      size: { width: '100%', height: detailsHeight },
    };
  }, [
    containerElement.clientHeight,
    containerElement.clientWidth,
    detailsHeight,
    detailsPosition,
    detailsWidth,
    dragStyles.dragHandleHorizontal,
    dragStyles.dragHandleVertical,
  ]);

  if (!showDetails.length) {
    return null;
  }

  return (
    <Resizable
      onResize={handleResize}
      handleClasses={resizable.handleClasses}
      defaultSize={resizable.defaultSize}
      size={resizable.size}
      enable={resizable.enable}
      minHeight={100}
      maxHeight={resizable.maxHeight}
      minWidth={180}
      maxWidth={resizable.maxWidth}
    >
      <div className={styles.container} ref={containerRef}>
        <div className={styles.scrollContainer}>
          <LogLineDetailsComponent log={showDetails[0]} logOptionsStorageKey={logOptionsStorageKey} logs={logs} />
        </div>
      </div>
    </Resizable>
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  container: css({
    overflow: 'auto',
    height: '100%',
    boxShadow: theme.shadows.z1,
    border: `1px solid ${theme.colors.border.medium}`,
    borderRight: 'none',
  }),
  scrollContainer: css({
    overflow: 'auto',
    height: '100%',
  }),
  componentWrapper: css({
    padding: theme.spacing(0, 1, 1, 1),
  }),
});
