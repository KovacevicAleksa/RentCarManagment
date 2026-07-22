import { useEffect, useRef } from "react";
import { Bell } from "lucide-react";

// NotificationBell shows the unread badge and rings whenever ringKey changes
// (i.e. a new notification arrives over the WebSocket). The ring is driven by
// restarting a CSS animation on the DOM node directly, so the effect only
// touches an external system and never calls setState.
export default function NotificationBell({ count, ringKey, onClick }) {
  const bellRef = useRef(null);
  const firstRender = useRef(true);

  useEffect(() => {
    if (firstRender.current) {
      firstRender.current = false;
      return;
    }
    const el = bellRef.current;
    if (!el) return;
    el.classList.remove("animate-bell-ring");
    void el.offsetWidth; // force reflow so the animation restarts every time
    el.classList.add("animate-bell-ring");
  }, [ringKey]);

  return (
    <button
      onClick={onClick}
      aria-label="Obaveštenja"
      className="relative flex items-center justify-center w-10 h-10 rounded-xl hover:bg-gray-100 transition-all"
    >
      <Bell ref={bellRef} className="w-5 h-5 text-gray-600" />
      {count > 0 && (
        <span className="absolute -top-0.5 -right-0.5 flex items-center justify-center min-w-[18px] h-[18px] px-1 bg-red-500 text-white text-[10px] font-bold rounded-full shadow">
          {count > 99 ? "99+" : count}
        </span>
      )}
    </button>
  );
}
