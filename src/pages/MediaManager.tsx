import React, { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Upload, Images } from 'lucide-react'
import MediaUpload from '@/components/MediaUpload'
import MediaGallery from '@/components/MediaGallery'
import type { MediaFile } from '@/lib/supabase'

const MediaManager = () => {
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  const handleUploadSuccess = (files: MediaFile[]) => {
    // Trigger gallery refresh
    setRefreshTrigger(prev => prev + 1)
  }

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground mb-2">Media Manager</h1>
          <p className="text-muted-foreground">
            Upload and manage your media files with Supabase Storage and PostgreSQL
          </p>
        </div>

        <Tabs defaultValue="upload" className="space-y-6">
          <TabsList className="grid w-[400px] grid-cols-2">
            <TabsTrigger value="upload" className="flex items-center space-x-2">
              <Upload className="w-4 h-4" />
              <span>Upload</span>
            </TabsTrigger>
            <TabsTrigger value="gallery" className="flex items-center space-x-2">
              <Images className="w-4 h-4" />
              <span>Gallery</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="upload">
            <MediaUpload onUploadSuccess={handleUploadSuccess} />
          </TabsContent>

          <TabsContent value="gallery">
            <MediaGallery refreshTrigger={refreshTrigger} />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}

export default MediaManager