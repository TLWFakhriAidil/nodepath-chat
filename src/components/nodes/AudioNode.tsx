import React, { useState, useRef } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Mic, Edit3, Upload, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function AudioNode({ data, id }: NodeProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [audioUrl, setAudioUrl] = useState((data?.audioUrl as string) || '');
  const [duration, setDuration] = useState((data?.duration as number) || 30);
  const [uploadedFile, setUploadedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file && file.type.startsWith('audio/')) {
      setUploadedFile(file);
      setAudioUrl(file.name);
      
      // Create object URL for preview
      const url = URL.createObjectURL(file);
      setPreviewUrl(url);
      
      // Convert to base64 for storage
      const reader = new FileReader();
      reader.onload = (e) => {
        const base64 = e.target?.result as string;
        // Update the node data immediately with base64
        (data?.onUpdate as Function)?.(id, {
          audioUrl: base64,
          mediaUrl: base64,
          duration,
          uploadedFile: {
            name: file.name,
            type: file.type,
            size: file.size
          },
          previewUrl: base64 // Store base64 instead of blob URL
        });
      };
      reader.readAsDataURL(file);
    }
  };

  // Initialize previewUrl from existing data
  React.useEffect(() => {
    if (data?.previewUrl && !previewUrl) {
      setPreviewUrl(data.previewUrl as string);
    }
  }, [data?.previewUrl]);

  // Cleanup object URL on unmount
  React.useEffect(() => {
    return () => {
      if (previewUrl && previewUrl.startsWith('blob:')) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  const handleSave = () => {
    setIsEditing(false);
    // Update the node data with current values
    (data?.onUpdate as Function)?.(id, {
      audioUrl,
      duration,
      uploadedFile: uploadedFile ? {
        name: uploadedFile.name,
        type: uploadedFile.type,
        size: uploadedFile.size
      } : null,
      previewUrl
    });
  };

  return (
    <div className="bg-card rounded-lg shadow-node border border-border min-w-[250px] max-w-[350px]">
      <Handle 
        type="target" 
        position={Position.Top} 
        className="w-3 h-3 bg-primary border-2 border-white"
      />
      
      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-green-500 mr-2" />
            <Mic className="w-4 h-4 text-green-500 mr-2" />
            <span className="text-sm font-medium text-black">Send Audio</span>
          </div>
          <div className="flex items-center gap-1">
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setIsEditing(!isEditing)}
              className="h-6 w-6 p-0"
            >
              <Edit3 className="w-3 h-3" />
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => (data?.onDelete as Function)?.(id)}
              className="h-6 w-6 p-0 text-destructive hover:text-destructive"
            >
              <Trash2 className="w-3 h-3" />
            </Button>
          </div>
        </div>
        
        {isEditing ? (
          <div className="space-y-3">
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Upload Audio</label>
              <input
                ref={fileInputRef}
                type="file"
                accept="audio/*"
                onChange={handleFileUpload}
                className="hidden"
              />
              <Button
                size="sm"
                variant="outline"
                onClick={() => fileInputRef.current?.click()}
                className="w-full"
              >
                <Upload className="w-4 h-4 mr-2" />
                Choose Audio
              </Button>
            </div>
            <div className="text-center text-xs text-muted-foreground">or</div>
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Audio URL</label>
              <Input
                value={audioUrl}
                onChange={(e) => setAudioUrl(e.target.value)}
                className="text-sm text-black bg-white border-gray-300"
                placeholder="https://example.com/audio.mp3"
              />
            </div>
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Duration (seconds)</label>
              <Input
                type="number"
                value={duration}
                onChange={(e) => setDuration(parseInt(e.target.value) || 0)}
                className="text-sm text-black bg-white border-gray-300"
                min="1"
                max="600"
              />
            </div>
            <Button size="sm" onClick={handleSave} className="w-full">
              Save
            </Button>
          </div>
        ) : (
          <div className="space-y-2">
            {(previewUrl || audioUrl) && (
              <div className="bg-muted/50 rounded p-2">
                {previewUrl ? (
                  <audio 
                    src={previewUrl} 
                    controls 
                    className="w-full"
                  />
                ) : audioUrl.startsWith('http') ? (
                  <audio 
                    src={audioUrl} 
                    controls 
                    className="w-full"
                  />
                ) : (
                  <div className="text-center">
                    <Upload className="w-8 h-8 mx-auto text-muted-foreground mb-1" />
                    <div className="text-xs text-muted-foreground truncate">{audioUrl}</div>
                  </div>
                )}
              </div>
            )}
            <div className="bg-muted/50 rounded p-3 text-center">
              <div className="text-2xl font-bold text-green-500">{duration}s</div>
              <div className="text-xs text-gray-600">audio duration</div>
            </div>
          </div>
        )}
      </div>
      
      <Handle 
        type="source" 
        position={Position.Bottom} 
        className="w-3 h-3 bg-primary border-2 border-white"
      />
    </div>
  );
}