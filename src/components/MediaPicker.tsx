import React, { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import { Upload, Link as LinkIcon, Image, Music, Video, Trash2 } from 'lucide-react'
import { supabase } from '@/integrations/supabase/client'
import { useToast } from '@/hooks/use-toast'

type MediaFile = {
  id: string;
  filename: string;
  original_name: string;
  file_type: string;
  file_size: number;
  storage_path: string;
  public_url: string;
  uploaded_at: string;
};

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
  const [mediaFiles, setMediaFiles] = useState<MediaFile[]>([])
  const { toast } = useToast()

  useEffect(() => {
    if (isOpen) {
      loadMediaFiles()
    }
  }, [isOpen])

  const loadMediaFiles = async () => {
    try {
      const { data, error } = await supabase
        .from('media_files')
        .select('*')
        .eq('file_type', mediaType)
        .order('uploaded_at', { ascending: false })

      if (error) throw error
      setMediaFiles(data || [])
    } catch (error) {
      console.error('Error loading media files:', error)
    }
  }

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
    const fileType = file.type.split('/')[0]
    if (fileType !== mediaType) {
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
      // Generate unique filename
      const fileExt = file.name.split('.').pop()
      const fileName = `${Date.now()}-${Math.random().toString(36).substring(2)}.${fileExt}`
      const filePath = `${fileName}`

      // Upload to Supabase Storage
      const { error: uploadError } = await supabase.storage
        .from('media')
        .upload(filePath, file)

      if (uploadError) throw uploadError

      // Get public URL
      const { data: { publicUrl } } = supabase.storage
        .from('media')
        .getPublicUrl(filePath)

      // Save metadata to database
      const { data, error: dbError } = await supabase
        .from('media_files')
        .insert({
          filename: fileName,
          original_name: file.name,
          file_type: fileType,
          file_size: file.size,
          storage_path: filePath,
          public_url: publicUrl
        })
        .select()
        .single()

      if (dbError) throw dbError

      onSelect(data.id)
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

  const handleDeleteMedia = async (mediaId: string) => {
    try {
      // Get the media file to find storage path
      const { data: mediaFile, error: fetchError } = await supabase
        .from('media_files')
        .select('storage_path')
        .eq('id', mediaId)
        .single()

      if (fetchError) throw fetchError

      // Delete from storage
      const { error: storageError } = await supabase.storage
        .from('media')
        .remove([mediaFile.storage_path])

      if (storageError) throw storageError

      // Delete from database
      const { error: dbError } = await supabase
        .from('media_files')
        .delete()
        .eq('id', mediaId)

      if (dbError) throw dbError

      // Refresh media files
      loadMediaFiles()
      toast({
        title: "Media deleted",
        description: "Media file has been removed"
      })
    } catch (error) {
      console.error('Error deleting media file:', error)
      toast({
        title: "Error",
        description: "Failed to delete media file",
        variant: "destructive"
      })
    }
  }

  const renderMediaPreview = (file: MediaFile) => {
    switch (mediaType) {
      case 'image':
        return (
          <img
            src={file.public_url}
            alt={file.original_name}
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
            src={file.public_url}
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
                        <p className="text-xs font-medium truncate" title={file.original_name}>
                          {file.original_name}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {(file.file_size / 1024 / 1024).toFixed(2)} MB
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