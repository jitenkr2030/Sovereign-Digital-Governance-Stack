#!/usr/bin/env python3
"""
Tests for Offline Bundle Generator

This module contains unit tests for the offline bundle generator
functionality.
"""

import json
import os
import shutil
import subprocess
import sys
import tarfile
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

from bundle_generator import (
    ContainerImage,
    HelmChart,
    ConfigBundle,
    BundleManifest,
    OfflineBundleGenerator,
    BundleVerifier
)


class TestContainerImage(unittest.TestCase):
    """Tests for ContainerImage class."""
    
    def test_full_name(self):
        """Test full_name property."""
        image = ContainerImage(
            repository="neam/core-api",
            tag="v1.0.0"
        )
        self.assertEqual(image.full_name, "neam/core-api:v1.0.0")
    
    def test_reference_with_digest(self):
        """Test reference property with digest."""
        image = ContainerImage(
            repository="neam/core-api",
            tag="v1.0.0",
            digest="sha256:abc123"
        )
        self.assertEqual(
            image.reference,
            "neam/core-api@sha256:abc123"
        )
    
    def test_reference_without_digest(self):
        """Test reference property without digest."""
        image = ContainerImage(
            repository="neam/core-api",
            tag="v1.0.0"
        )
        self.assertEqual(image.reference, "neam/core-api:v1.0.0")
    
    def test_to_dict(self):
        """Test to_dict method."""
        image = ContainerImage(
            repository="neam/core-api",
            tag="v1.0.0",
            digest="sha256:abc123",
            platform="linux/arm64"
        )
        data = image.to_dict()
        
        self.assertEqual(data["repository"], "neam/core-api")
        self.assertEqual(data["tag"], "v1.0.0")
        self.assertEqual(data["digest"], "sha256:abc123")
        self.assertEqual(data["platform"], "linux/arm64")
    
    def test_from_dict(self):
        """Test from_dict class method."""
        data = {
            "repository": "neam/core-api",
            "tag": "v1.0.0",
            "digest": "sha256:abc123",
            "platform": "linux/arm64"
        }
        image = ContainerImage.from_dict(data)
        
        self.assertEqual(image.repository, "neam/core-api")
        self.assertEqual(image.tag, "v1.0.0")
        self.assertEqual(image.digest, "sha256:abc123")
        self.assertEqual(image.platform, "linux/arm64")


class TestHelmChart(unittest.TestCase):
    """Tests for HelmChart class."""
    
    def test_to_dict(self):
        """Test to_dict method."""
        chart = HelmChart(
            name="neam-core",
            version="1.0.0",
            repository="neam-charts",
            chart_url="https://charts.neam.internal/neam-core-1.0.0.tgz",
            values_file="values-production.yaml",
            dependencies=["neam-common-1.0.0"]
        )
        data = chart.to_dict()
        
        self.assertEqual(data["name"], "neam-core")
        self.assertEqual(data["version"], "1.0.0")
        self.assertEqual(data["repository"], "neam-charts")
        self.assertEqual(data["chart_url"], "https://charts.neam.internal/neam-core-1.0.0.tgz")
        self.assertEqual(data["values_file"], "values-production.yaml")
        self.assertEqual(data["dependencies"], ["neam-common-1.0.0"])
    
    def test_from_dict(self):
        """Test from_dict class method."""
        data = {
            "name": "neam-core",
            "version": "1.0.0",
            "repository": "neam-charts",
            "chart_url": "https://charts.neam.internal/neam-core-1.0.0.tgz",
            "values_file": "values-production.yaml",
            "dependencies": ["neam-common-1.0.0"]
        }
        chart = HelmChart.from_dict(data)
        
        self.assertEqual(chart.name, "neam-core")
        self.assertEqual(chart.version, "1.0.0")
        self.assertEqual(chart.dependencies, ["neam-common-1.0.0"])


class TestConfigBundle(unittest.TestCase):
    """Tests for ConfigBundle class."""
    
    def test_to_dict(self):
        """Test to_dict method."""
        config = ConfigBundle(
            name="production-config",
            source_path="k8s/production",
            config_type="kubernetes",
            includes=["secrets", "configmaps"],
            transformers=["namespace-isolation", "image-registry"]
        )
        data = config.to_dict()
        
        self.assertEqual(data["name"], "production-config")
        self.assertEqual(data["source_path"], "k8s/production")
        self.assertEqual(data["config_type"], "kubernetes")
        self.assertEqual(len(data["includes"]), 2)
        self.assertEqual(len(data["transformers"]), 2)
    
    def test_from_dict(self):
        """Test from_dict class method."""
        data = {
            "name": "production-config",
            "source_path": "k8s/production",
            "config_type": "kubernetes",
            "includes": ["secrets"],
            "transformers": ["namespace-isolation"]
        }
        config = ConfigBundle.from_dict(data)
        
        self.assertEqual(config.name, "production-config")
        self.assertEqual(config.transformers, ["namespace-isolation"])


class TestBundleManifest(unittest.TestCase):
    """Tests for BundleManifest class."""
    
    def test_to_dict(self):
        """Test to_dict method."""
        manifest = BundleManifest(
            bundle_version="1.0.0",
            bundle_type="complete",
            created_at="2024-01-01T00:00:00",
            platform_version="1.0.0",
            target_environment="air-gapped"
        )
        manifest.container_images = [{"repository": "test", "tag": "v1"}]
        manifest.total_size = 1024
        manifest.checksum = "abc123"
        
        data = manifest.to_dict()
        
        self.assertEqual(data["bundle_version"], "1.0.0")
        self.assertEqual(data["container_images"][0]["repository"], "test")
        self.assertEqual(data["total_size"], 1024)
    
    def test_save_and_load(self):
        """Test save and load methods."""
        manifest = BundleManifest(
            bundle_version="1.0.0",
            bundle_type="complete",
            created_at="2024-01-01T00:00:00",
            platform_version="1.0.0",
            target_environment="air-gapped"
        )
        manifest.checksum = "test-checksum"
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            temp_path = f.name
        
        try:
            manifest.save(temp_path)
            loaded = BundleManifest.load(temp_path)
            
            self.assertEqual(loaded.bundle_version, "1.0.0")
            self.assertEqual(loaded.checksum, "test-checksum")
        finally:
            os.unlink(temp_path)


class TestOfflineBundleGenerator(unittest.TestCase):
    """Tests for OfflineBundleGenerator class."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.temp_dir = tempfile.mkdtemp()
        self.output_dir = os.path.join(self.temp_dir, "output")
        os.makedirs(self.output_dir)
    
    def tearDown(self):
        """Clean up test fixtures."""
        shutil.rmtree(self.temp_dir, ignore_errors=True)
    
    @patch('bundle_generator.shutil.which')
    def test_check_prerequisites_missing(self, mock_which):
        """Test prerequisite check with missing tools."""
        mock_which.return_value = None
        
        with self.assertRaises(RuntimeError) as context:
            OfflineBundleGenerator(self.output_dir)
        
        self.assertIn("Missing required tools", str(context.exception))
    
    @patch('bundle_generator.shutil.which')
    def test_check_prerequisites_satisfied(self, mock_which):
        """Test prerequisite check when all tools are present."""
        mock_which.return_value = "/usr/bin/docker"
        
        generator = OfflineBundleGenerator(self.output_dir)
        self.assertIsNotNone(generator)
    
    def test_add_container_images_skip_existing(self):
        """Test skipping existing images."""
        with patch('bundle_generator.shutil.which'):
            generator = OfflineBundleGenerator(self.output_dir)
            generator.bundle_dir = Path(self.output_dir) / "bundle"
            generator.bundle_dir.mkdir(parents=True)
            
            image = ContainerImage(
                repository="test/repo",
                tag="v1.0.0"
            )
            
            # Add image first time
            generator.manifest.container_images.append(image.to_dict())
            
            # Try to add same image again (should skip)
            with patch.object(generator, '_export_container_image') as mock_export:
                count = generator.add_container_images([image], skip_existing=True)
                self.assertEqual(count, 0)
                mock_export.assert_not_called()
    
    def test_generate_manifest(self):
        """Test manifest generation."""
        with patch('bundle_generator.shutil.which'):
            generator = OfflineBundleGenerator(self.output_dir)
            generator.bundle_dir = Path(self.output_dir) / "bundle"
            generator.bundle_dir.mkdir(parents=True)
            
            # Create some files
            (generator.bundle_dir / "test.txt").write_text("test content")
            
            manifest = generator.generate_manifest()
            
            self.assertIsNotNone(manifest)
            self.assertEqual(manifest.platform_version, "1.0.0")
            self.assertEqual(manifest.target_environment, "air-gapped")
            self.assertGreater(manifest.total_size, 0)
            self.assertIsNotNone(manifest.checksum)
    
    def test_format_size(self):
        """Test size formatting."""
        with patch('bundle_generator.shutil.which'):
            generator = OfflineBundleGenerator(self.output_dir)
            
            self.assertEqual(generator._format_size(512), "512.00 B")
            self.assertEqual(generator._format_size(1024), "1.00 KB")
            self.assertEqual(generator._format_size(1024 * 1024), "1.00 MB")
            self.assertEqual(generator._format_size(1024 * 1024 * 1024), "1.00 GB")


class TestBundleVerifier(unittest.TestCase):
    """Tests for BundleVerifier class."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.temp_dir = tempfile.mkdtemp()
    
    def tearDown(self):
        """Clean up test fixtures."""
        shutil.rmtree(self.temp_dir, ignore_errors=True)
    
    def test_verify_nonexistent_bundle(self):
        """Test verification of nonexistent bundle."""
        verifier = BundleVerifier("/nonexistent/bundle.tar.gz")
        success = verifier.verify()
        
        self.assertFalse(success)
        self.assertIn("Bundle not found", verifier.errors[0])
    
    def test_verify_missing_manifest(self):
        """Test verification with missing manifest."""
        bundle_dir = Path(self.temp_dir) / "bundle"
        bundle_dir.mkdir()
        
        verifier = BundleVerifier(str(bundle_dir))
        success = verifier.verify()
        
        self.assertFalse(success)
        self.assertIn("Manifest not found", verifier.errors[0])
    
    def test_verify_checksum_mismatch(self):
        """Test verification with checksum mismatch."""
        # Create bundle directory
        bundle_dir = Path(self.temp_dir) / "bundle"
        bundle_dir.mkdir()
        
        # Create manifest with wrong checksum
        manifest = BundleManifest(
            bundle_version="1.0.0",
            bundle_type="complete",
            created_at="2024-01-01T00:00:00",
            platform_version="1.0.0",
            target_environment="air-gapped"
        )
        manifest.checksum = "wrong-checksum"
        manifest.save(str(bundle_dir / "bundle-manifest.yaml"))
        
        # Create test file
        (bundle_dir / "test.txt").write_text("test content")
        
        verifier = BundleVerifier(str(bundle_dir))
        verifier.manifest = manifest
        
        success = verifier.verify()
        
        self.assertFalse(success)
        self.assertTrue(any("Checksum mismatch" in e for e in verifier.errors))


class TestIntegration(unittest.TestCase):
    """Integration tests for bundle generator."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.temp_dir = tempfile.mkdtemp()
        self.output_dir = os.path.join(self.temp_dir, "output")
        os.makedirs(self.output_dir)
    
    def tearDown(self):
        """Clean up test fixtures."""
        shutil.rmtree(self.temp_dir, ignore_errors=True)
    
    @patch('bundle_generator.shutil.which')
    def test_create_bundle_with_images(self, mock_which):
        """Test creating a bundle with container images."""
        mock_which.return_value = "/usr/bin/docker"
        
        generator = OfflineBundleGenerator(
            self.output_dir,
            platform_version="1.0.0"
        )
        
        # Add test images (mocked export)
        image = ContainerImage(
            repository="test/repo",
            tag="v1.0.0"
        )
        
        with patch.object(generator, '_export_container_image') as mock_export:
            mock_export.return_value = os.path.join(
                self.output_dir, "test.tar"
            )
            (Path(self.output_dir) / "test.tar").touch()
            
            generator.add_container_images([image], skip_existing=False)
        
        # Create bundle
        bundle_path = generator.create_bundle()
        
        self.assertTrue(os.path.exists(bundle_path))
        self.assertTrue(bundle_path.endswith(".tar.gz"))
        
        # Verify bundle
        verifier = BundleVerifier(bundle_path)
        success = verifier.verify()
        
        self.assertTrue(success)
    
    @patch('bundle_generator.shutil.which')
    def test_create_bundle_with_charts(self, mock_which):
        """Test creating a bundle with Helm charts."""
        mock_which.return_value = "/usr/bin/helm"
        
        generator = OfflineBundleGenerator(
            self.output_dir,
            platform_version="1.0.0"
        )
        
        # Add test chart (mocked export)
        chart = HelmChart(
            name="test-chart",
            version="1.0.0",
            repository="test-repo",
            chart_url="https://example.com/test-chart-1.0.0.tgz"
        )
        
        with patch.object(generator, '_export_helm_chart') as mock_export:
            mock_export.return_value = os.path.join(
                self.output_dir, "test-chart-1.0.0.tgz"
            )
            (Path(self.output_dir) / "test-chart-1.0.0.tgz").touch()
            
            generator.add_helm_charts([chart], skip_existing=False)
        
        # Create bundle
        bundle_path = generator.create_bundle()
        
        self.assertTrue(os.path.exists(bundle_path))
        
        # Verify bundle
        verifier = BundleVerifier(bundle_path)
        success = verifier.verify()
        
        self.assertTrue(success)
    
    @patch('bundle_generator.shutil.which')
    def test_create_bundle_with_configs(self, mock_which):
        """Test creating a bundle with configuration bundles."""
        mock_which.return_value = "/usr/bin/kubectl"
        
        generator = OfflineBundleGenerator(
            self.output_dir,
            platform_version="1.0.0"
        )
        
        # Create test config directory
        config_source = Path(self.temp_dir) / "config-source"
        config_source.mkdir()
        (config_source / "deployment.yaml").write_text("apiVersion: v1")
        
        # Add config bundle
        config = ConfigBundle(
            name="test-config",
            source_path="config-source",
            config_type="kubernetes"
        )
        generator.add_config_bundle(config, base_path=self.temp_dir)
        
        # Create bundle
        bundle_path = generator.create_bundle()
        
        self.assertTrue(os.path.exists(bundle_path))
        
        # Verify bundle
        verifier = BundleVerifier(bundle_path)
        success = verifier.verify()
        
        self.assertTrue(success)


if __name__ == "__main__":
    unittest.main(verbosity=2)
