import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { 
  Upload, 
  Search, 
  Filter, 
  Grid3X3, 
  List, 
  Download, 
  Trash2, 
  Eye, 
  Edit,
  Image,
  FileText,
  Video,
  Music,
  File,
  FolderPlus,
  MoreVertical,
  Star,
  Share2
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface MediaFile {
  id: string;
  name: string;
  type: 'image' | 'video' | 'audio' | 'document' | 'other';
  size: string;
  uploadDate: Date;
  url: string;
  thumbnail?: string;
  starred: boolean;
}

const MediaManager = () => {
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
  const [filterType, setFilterType] = useState<string>('all');

  // Mock data
  const [mediaFiles] = useState<MediaFile[]>([
    {
      id: '1',
      name: 'chatbot-avatar.png',
      type: 'image',
      size: '245 KB',
      uploadDate: new Date('2024-01-15'),
      url: '/placeholder-image.jpg',
      starred: true
    },
    {
      id: '2',
      name: 'flow-diagram.pdf',
      type: 'document',
      size: '1.2 MB',
      uploadDate: new Date('2024-01-14'),
      url: '/placeholder-doc.pdf',
      starred: false
    },
    {
      id: '3',
      name: 'demo-video.mp4',
      type: 'video',
      size: '15.8 MB',
      uploadDate: new Date('2024-01-13'),
      url: '/placeholder-video.mp4',
      starred: false
    },
    {
      id: '4',
      name: 'notification-sound.mp3',
      type: 'audio',
      size: '892 KB',
      uploadDate: new Date('2024-01-12'),
      url: '/placeholder-audio.mp3',
      starred: true
    },
    {
      id: '5',
      name: 'user-guide.docx',
      type: 'document',
      size: '567 KB',
      uploadDate: new Date('2024-01-11'),
      url: '/placeholder-doc.docx',
      starred: false
    }
  ]);

  const getFileIcon = (type: string) => {
    switch (type) {
      case 'image': return Image;
      case 'video': return Video;
      case 'audio': return Music;
      case 'document': return FileText;
      default: return File;
    }
  };

  const getFileTypeColor = (type: string) => {
    switch (type) {
      case 'image': return 'bg-green-100 text-green-700 dark:bg-green-900/20 dark:text-green-400';
      case 'video': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400';
      case 'audio': return 'bg-purple-100 text-purple-700 dark:bg-purple-900/20 dark:text-purple-400';
      case 'document': return 'bg-orange-100 text-orange-700 dark:bg-orange-900/20 dark:text-orange-400';
      default: return 'bg-gray-100 text-gray-700 dark:bg-gray-900/20 dark:text-gray-400';
    }
  };

  const filteredFiles = mediaFiles.filter(file => {
    const matchesSearch = file.name.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesFilter = filterType === 'all' || file.type === filterType;
    return matchesSearch && matchesFilter;
  });

  const handleFileSelect = (fileId: string) => {
    setSelectedFiles(prev => 
      prev.includes(fileId) 
        ? prev.filter(id => id !== fileId)
        : [...prev, fileId]
    );
  };

  const handleSelectAll = () => {
    if (selectedFiles.length === filteredFiles.length) {
      setSelectedFiles([]);
    } else {
      setSelectedFiles(filteredFiles.map(file => file.id));
    }
  };

  const fileTypeStats = {
    all: mediaFiles.length,
    image: mediaFiles.filter(f => f.type === 'image').length,
    video: mediaFiles.filter(f => f.type === 'video').length,
    audio: mediaFiles.filter(f => f.type === 'audio').length,
    document: mediaFiles.filter(f => f.type === 'document').length,
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-2">
            Media Manager
          </h1>
          <p className="text-slate-600 dark:text-slate-400">
            Manage your media files, images, documents, and assets
          </p>
        </div>
        
        <div className="flex items-center space-x-2">
          <Button className="bg-blue-600 hover:bg-blue-700">
            <Upload className="w-4 h-4 mr-2" />
            Upload Files
          </Button>
          <Button variant="outline">
            <FolderPlus className="w-4 h-4 mr-2" />
            New Folder
          </Button>
        </div>
      </div>

      {/* Toolbar */}
      <Card className="border-0 shadow-lg">
        <CardContent className="p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              {/* Search */}
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400 w-4 h-4" />
                <Input
                  placeholder="Search files..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 w-64"
                />
              </div>
              
              {/* Filter */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="sm">
                    <Filter className="w-4 h-4 mr-2" />
                    {filterType === 'all' ? 'All Files' : filterType.charAt(0).toUpperCase() + filterType.slice(1)}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem onClick={() => setFilterType('all')}>
                    All Files ({fileTypeStats.all})
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setFilterType('image')}>
                    Images ({fileTypeStats.image})
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setFilterType('video')}>
                    Videos ({fileTypeStats.video})
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setFilterType('audio')}>
                    Audio ({fileTypeStats.audio})
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setFilterType('document')}>
                    Documents ({fileTypeStats.document})
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
              
              {selectedFiles.length > 0 && (
                <div className="flex items-center space-x-2">
                  <Badge variant="secondary">
                    {selectedFiles.length} selected
                  </Badge>
                  <Button variant="outline" size="sm">
                    <Download className="w-4 h-4 mr-2" />
                    Download
                  </Button>
                  <Button variant="outline" size="sm">
                    <Trash2 className="w-4 h-4 mr-2" />
                    Delete
                  </Button>
                </div>
              )}
            </div>
            
            <div className="flex items-center space-x-2">
              <Button
                variant={viewMode === 'grid' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('grid')}
              >
                <Grid3X3 className="w-4 h-4" />
              </Button>
              <Button
                variant={viewMode === 'list' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('list')}
              >
                <List className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* File Grid/List */}
      <Card className="border-0 shadow-xl">
        <CardContent className="p-6">
          {filteredFiles.length === 0 ? (
            <div className="text-center py-12">
              <Upload className="w-12 h-12 text-slate-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 dark:text-white mb-2">
                No files found
              </h3>
              <p className="text-slate-600 dark:text-slate-400 mb-4">
                {searchQuery ? 'Try adjusting your search terms' : 'Upload your first file to get started'}
              </p>
              <Button className="bg-blue-600 hover:bg-blue-700">
                <Upload className="w-4 h-4 mr-2" />
                Upload Files
              </Button>
            </div>
          ) : viewMode === 'grid' ? (
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
              {filteredFiles.map((file) => {
                const FileIcon = getFileIcon(file.type);
                const isSelected = selectedFiles.includes(file.id);
                
                return (
                  <div
                    key={file.id}
                    className={`relative group cursor-pointer rounded-lg border-2 transition-all ${
                      isSelected 
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' 
                        : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'
                    }`}
                    onClick={() => handleFileSelect(file.id)}
                  >
                    <div className="p-4">
                      {/* File Icon/Thumbnail */}
                      <div className="w-full h-24 bg-slate-100 dark:bg-slate-800 rounded-lg flex items-center justify-center mb-3">
                        {file.type === 'image' ? (
                          <div className="w-full h-full bg-gradient-to-br from-blue-400 to-purple-500 rounded-lg flex items-center justify-center">
                            <Image className="w-8 h-8 text-white" />
                          </div>
                        ) : (
                          <FileIcon className="w-8 h-8 text-slate-400" />
                        )}
                      </div>
                      
                      {/* File Info */}
                      <div className="space-y-1">
                        <p className="text-sm font-medium text-slate-900 dark:text-white truncate">
                          {file.name}
                        </p>
                        <div className="flex items-center justify-between">
                          <Badge className={`text-xs ${getFileTypeColor(file.type)}`}>
                            {file.type}
                          </Badge>
                          <span className="text-xs text-slate-500">{file.size}</span>
                        </div>
                      </div>
                      
                      {/* Actions */}
                      <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                              <MoreVertical className="w-4 h-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem>
                              <Eye className="w-4 h-4 mr-2" />
                              Preview
                            </DropdownMenuItem>
                            <DropdownMenuItem>
                              <Download className="w-4 h-4 mr-2" />
                              Download
                            </DropdownMenuItem>
                            <DropdownMenuItem>
                              <Edit className="w-4 h-4 mr-2" />
                              Rename
                            </DropdownMenuItem>
                            <DropdownMenuItem>
                              <Share2 className="w-4 h-4 mr-2" />
                              Share
                            </DropdownMenuItem>
                            <DropdownMenuItem className="text-red-600">
                              <Trash2 className="w-4 h-4 mr-2" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                      
                      {/* Star */}
                      {file.starred && (
                        <div className="absolute top-2 left-2">
                          <Star className="w-4 h-4 text-yellow-500 fill-current" />
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="space-y-2">
              {/* List Header */}
              <div className="grid grid-cols-12 gap-4 p-3 text-sm font-medium text-slate-600 dark:text-slate-400 border-b">
                <div className="col-span-1">
                  <input
                    type="checkbox"
                    checked={selectedFiles.length === filteredFiles.length && filteredFiles.length > 0}
                    onChange={handleSelectAll}
                    className="rounded"
                  />
                </div>
                <div className="col-span-5">Name</div>
                <div className="col-span-2">Type</div>
                <div className="col-span-2">Size</div>
                <div className="col-span-2">Modified</div>
              </div>
              
              {/* List Items */}
              {filteredFiles.map((file) => {
                const FileIcon = getFileIcon(file.type);
                const isSelected = selectedFiles.includes(file.id);
                
                return (
                  <div
                    key={file.id}
                    className={`grid grid-cols-12 gap-4 p-3 rounded-lg transition-colors ${
                      isSelected 
                        ? 'bg-blue-50 dark:bg-blue-900/20' 
                        : 'hover:bg-slate-50 dark:hover:bg-slate-800/50'
                    }`}
                  >
                    <div className="col-span-1 flex items-center">
                      <input
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => handleFileSelect(file.id)}
                        className="rounded"
                      />
                    </div>
                    <div className="col-span-5 flex items-center space-x-3">
                      <FileIcon className="w-5 h-5 text-slate-400" />
                      <div className="flex items-center space-x-2">
                        <span className="text-sm font-medium text-slate-900 dark:text-white">
                          {file.name}
                        </span>
                        {file.starred && (
                          <Star className="w-4 h-4 text-yellow-500 fill-current" />
                        )}
                      </div>
                    </div>
                    <div className="col-span-2 flex items-center">
                      <Badge className={`text-xs ${getFileTypeColor(file.type)}`}>
                        {file.type}
                      </Badge>
                    </div>
                    <div className="col-span-2 flex items-center">
                      <span className="text-sm text-slate-600 dark:text-slate-400">{file.size}</span>
                    </div>
                    <div className="col-span-2 flex items-center justify-between">
                      <span className="text-sm text-slate-600 dark:text-slate-400">
                        {file.uploadDate.toLocaleDateString()}
                      </span>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                            <MoreVertical className="w-4 h-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem>
                            <Eye className="w-4 h-4 mr-2" />
                            Preview
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Download className="w-4 h-4 mr-2" />
                            Download
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Edit className="w-4 h-4 mr-2" />
                            Rename
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Share2 className="w-4 h-4 mr-2" />
                            Share
                          </DropdownMenuItem>
                          <DropdownMenuItem className="text-red-600">
                            <Trash2 className="w-4 h-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Storage Info */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-700 dark:text-slate-300">Storage Used</p>
                <p className="text-2xl font-bold text-slate-900 dark:text-white">18.2 MB</p>
              </div>
              <div className="w-12 h-12 bg-blue-100 dark:bg-blue-900/20 rounded-lg flex items-center justify-center">
                <Upload className="w-6 h-6 text-blue-600" />
              </div>
            </div>
            <div className="mt-3">
              <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
                <div className="bg-blue-600 h-2 rounded-full" style={{ width: '18%' }} />
              </div>
              <p className="text-xs text-slate-500 mt-1">18% of 100 MB used</p>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-700 dark:text-slate-300">Total Files</p>
                <p className="text-2xl font-bold text-slate-900 dark:text-white">{mediaFiles.length}</p>
              </div>
              <div className="w-12 h-12 bg-green-100 dark:bg-green-900/20 rounded-lg flex items-center justify-center">
                <File className="w-6 h-6 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-700 dark:text-slate-300">Starred Files</p>
                <p className="text-2xl font-bold text-slate-900 dark:text-white">
                  {mediaFiles.filter(f => f.starred).length}
                </p>
              </div>
              <div className="w-12 h-12 bg-yellow-100 dark:bg-yellow-900/20 rounded-lg flex items-center justify-center">
                <Star className="w-6 h-6 text-yellow-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default MediaManager;