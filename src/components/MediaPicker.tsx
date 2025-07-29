import React, { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import { Upload, Link as LinkIcon, Image, Music, Video, Trash2 } from 'lucide-react'
import { MediaFile } from '@/types/chatbot'
import { getMediaFiles, createMediaFileFromFile, saveMediaFile, deleteMediaFile } from '@/lib/localStorage'
import { useToast } from '@/hooks/use-toast'

interface MediaPickerProps {
  mediaType: 'image' | 'audio' | 'video'
  selectedMediaId?: string
  onSelect: (mediaId?: string, mediaUrl?: string) => void
  children: React.ReactNode
}

const MediaPicker: React.FC<MediaPickerProps> = ({ 
  mediaType, 
  selectedMediaId, 
  onSelect, 
  children 
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const [mediaUrl, setMediaUrl] = useState('')
  const [uploading, setUploading] = useState(false)
  const { toast } = useToast()

  const mediaFiles = getMediaFiles().filter(file => 
    file.type.startsWith(mediaType + '/')
  )

  const getAcceptedTypes = () => {
    switch (mediaType) {
      case 'image':
        return '.png,.jpg,.jpeg'
      case 'audio':
        return '.mp3,.wav'
      case 'video':
        return '.mp4'
      default:
        return ''
    }
  }

  const getIcon = () => {
    switch (mediaType) {
      case 'image':
        return <Image className="w-4 h-4" />
      case 'audio':
        return <Music className="w-4 h-4" />
      case 'video':
        return <Video className="w-4 h-4" />
    }
  }

  const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    // Validate file type
    if (!file.type.startsWith(mediaType + '/')) {
      toast({
        title: "Invalid file type",
        description: `Please select a ${mediaType} file`,
        variant: "destructive"
      })
      return
    }

    // Validate file size (25MB max)
    if (file.size > 25 * 1024 * 1024) {
      toast({
        title: "File too large",
        description: "Please select a file smaller than 25MB",
        variant: "destructive"
      })
      return
    }

    setUploading(true)
    try {
      const mediaFile = await createMediaFileFromFile(file)
      saveMediaFile(mediaFile)
      onSelect(mediaFile.id)
      setIsOpen(false)
      
      toast({
        title: "Upload successful",
        description: `${file.name} has been uploaded`
      })
    } catch (error) {
      toast({
        title: "Upload failed",
        description: error instanceof Error ? error.message : 'Unknown error',
        variant: "destructive"
      })
    } finally {
      setUploading(false)
      // Reset input
      event.target.value = ''
    }
  }

  const handleUrlSubmit = () => {
    if (!mediaUrl.trim()) return
    
    onSelect(undefined, mediaUrl.trim())
    setMediaUrl('')
    setIsOpen(false)
    
    toast({
      title: "URL added",
      description: "Media URL has been set"
    })
  }

  const handleSelectMedia = (mediaId: string) => {
    onSelect(mediaId)
    setIsOpen(false)
  }

  const handleDeleteMedia = (mediaId: string) => {
    deleteMediaFile(mediaId)
    toast({
      title: "Media deleted",
      description: "Media file has been removed"
    })
  }

  const renderMediaPreview = (file: MediaFile) => {
    switch (mediaType) {
      case 'image':
        return (
          <img
            src={file.dataUrl}
            alt={file.filename}
            className="w-full h-24 object-cover rounded-md"
          />
        )
      case 'audio':
        return (
          <div className="w-full h-24 bg-muted rounded-md flex items-center justify-center">
            <Music className="w-8 h-8 text-muted-foreground" />
          </div>
        )
      case 'video':
        return (
          <video
            src={file.dataUrl}
            className="w-full h-24 object-cover rounded-md"
            muted
          />
        )
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {getIcon()}
            Select {mediaType.charAt(0).toUpperCase() + mediaType.slice(1)}
          </DialogTitle>
        </DialogHeader>
        
        <Tabs defaultValue="upload" className="w-full">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="upload">Upload File</TabsTrigger>
            <TabsTrigger value="url">Use URL</TabsTrigger>
            <TabsTrigger value="library">Media Library</TabsTrigger>
          </TabsList>
          
          <TabsContent value="upload" className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="file-upload">Upload {mediaType} file</Label>
              <div className="border-2 border-dashed border-muted-foreground/25 rounded-lg p-6 text-center">
                <input
                  id="file-upload"
                  type="file"
                  accept={getAcceptedTypes()}
                  onChange={handleFileUpload}
                  disabled={uploading}
                  className="hidden"
                />
                <label
                  htmlFor="file-upload"
                  className="cursor-pointer flex flex-col items-center gap-2"
                >
                  <Upload className="w-8 h-8 text-muted-foreground" />
                  <div>
                    <p className="font-medium">
                      {uploading ? 'Uploading...' : `Click to upload ${mediaType}`}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      Max size: 25MB
                    </p>
                  </div>
                </label>
              </div>
            </div>
          </TabsContent>
          
          <TabsContent value="url" className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="media-url">{mediaType.charAt(0).toUpperCase() + mediaType.slice(1)} URL</Label>
              <Input
                id="media-url"
                value={mediaUrl}
                onChange={(e) => setMediaUrl(e.target.value)}
                placeholder={`Enter ${mediaType} URL...`}
              />
            </div>
            <Button onClick={handleUrlSubmit} disabled={!mediaUrl.trim()}>
              <LinkIcon className="w-4 h-4 mr-2" />
              Use URL
            </Button>
          </TabsContent>
          
          <TabsContent value="library" className="space-y-4">
            {mediaFiles.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No {mediaType} files in your library. Upload some files first.
              </div>
            ) : (
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {mediaFiles.map((file) => (
                  <Card key={file.id} className="cursor-pointer hover:bg-muted/50">
                    <CardContent className="p-3">
                      <div className="relative">
                        {renderMediaPreview(file)}
                        <Button
                          variant="destructive"
                          size="sm"
                          className="absolute top-1 right-1 w-6 h-6 p-0"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleDeleteMedia(file.id)
                          }}
                        >
                          <Trash2 className="w-3 h-3" />
                        </Button>
                      </div>
                      <div className="mt-2">
                        <p className="text-xs font-medium truncate" title={file.filename}>
                          {file.filename}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {(file.size / 1024 / 1024).toFixed(2)} MB
                        </p>
                        <Button
                          variant="outline"
                          size="sm"
                          className="w-full mt-2"
                          onClick={() => handleSelectMedia(file.id)}
                        >
                          Select
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}

export default MediaPicker